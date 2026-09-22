package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type securityConfig struct {
	secret       []byte
	authAttempts int
	authWindow   time.Duration
	uploadBytes  int64
}

var security = securityConfig{authAttempts: 10, authWindow: time.Minute, uploadBytes: 201 << 20}

func loadSecurityConfig() (securityConfig, error) {
	c := securityConfig{authAttempts: 10, authWindow: time.Minute, uploadBytes: 201 << 20}
	mode := os.Getenv("APP_ENV")
	if mode != "" && mode != "production" && mode != "development" && mode != "test" {
		return c, errors.New("invalid APP_ENV")
	}
	raw := os.Getenv("JWT_SECRET")
	if mode == "development" && raw == "" {
		c.secret = []byte("clipsense-explicit-local-development-only")
	} else {
		decoded, err := base64.StdEncoding.Strict().DecodeString(raw)
		if err != nil || len(decoded) < 32 || len(decoded) > 64 || bytes.Equal(decoded, bytes.Repeat(decoded[:min(1, len(decoded))], len(decoded))) {
			return c, errors.New("JWT_SECRET must encode 32-64 random bytes as base64")
		}
		c.secret = decoded
	}
	for _, field := range []struct {
		name   string
		target *int
		max    int
	}{
		{"AUTH_MAX_ATTEMPTS", &c.authAttempts, 1000},
	} {
		if raw, exists := os.LookupEnv(field.name); exists {
			n, err := strconv.Atoi(raw)
			if err != nil || n < 1 || n > field.max {
				return c, fmt.Errorf("invalid %s", field.name)
			}
			*field.target = n
		}
	}
	if raw, exists := os.LookupEnv("AUTH_WINDOW_SECONDS"); exists {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > 3600 {
			return c, errors.New("invalid AUTH_WINDOW_SECONDS")
		}
		c.authWindow = time.Duration(n) * time.Second
	}
	if raw, exists := os.LookupEnv("MAX_UPLOAD_REQUEST_BYTES"); exists {
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || n < 1024 || n > 1<<30 {
			return c, errors.New("invalid MAX_UPLOAD_REQUEST_BYTES")
		}
		c.uploadBytes = n
	}
	return c, nil
}

func publicError(w http.ResponseWriter, status int, code string) {
	recordRejection(status, code)
	writeJSONStatus(w, status, map[string]string{"error": code})
}

// Fixed secret-keyed buckets bound Redis memory even with many IPv6 addresses.
// Peer IP is authoritative; no client-supplied forwarding header is trusted.
func authBucket(remote string) string {
	host, _, err := net.SplitHostPort(remote)
	if err != nil {
		host = remote
	}
	if ip := net.ParseIP(host); ip != nil {
		host = ip.String()
	}
	h := hmac.New(sha256.New, jwtSecret)
	h.Write([]byte(host))
	return fmt.Sprintf("auth:attempts:%04x", binary.BigEndian.Uint16(h.Sum(nil)[:2]))
}

var attemptScript = redis.NewScript(`
local count = redis.call('INCR', KEYS[1])
if count == 1 then redis.call('PEXPIRE', KEYS[1], ARGV[1]) end
return count
`)

func authAttemptLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()
		count, err := attemptScript.Run(ctx, rdb, []string{authBucket(r.RemoteAddr)}, security.authWindow.Milliseconds()).Int()
		if err != nil {
			publicError(w, 503, "authentication_unavailable")
			return
		}
		if count > security.authAttempts {
			w.Header().Set("Retry-After", strconv.Itoa(int(security.authWindow.Seconds())))
			publicError(w, 429, "authentication_rate_limited")
			return
		}
		next.ServeHTTP(w, r)
	})
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func readCredentials(w http.ResponseWriter, r *http.Request) (credentials, bool) {
	var c credentials
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&c); err != nil {
		publicError(w, 400, "invalid_credentials_request")
		return c, false
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		publicError(w, 400, "invalid_credentials_request")
		return c, false
	}
	c.Email = strings.TrimSpace(c.Email)
	address, err := mail.ParseAddress(c.Email)
	if err != nil || address.Address != c.Email || len(c.Email) > 254 || !strings.Contains(c.Email, "@") || len(c.Password) < 8 || len(c.Password) > 72 || strings.TrimSpace(c.Password) == "" {
		publicError(w, 400, "invalid_credentials_request")
		return c, false
	}
	return c, true
}

func register(w http.ResponseWriter, r *http.Request) {
	c, ok := readCredentials(w, r)
	if !ok {
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(c.Password), bcrypt.DefaultCost)
	if err != nil {
		publicError(w, 500, "authentication_failed")
		return
	}
	id := uuid.NewString()
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	_, err = db.ExecContext(ctx, `INSERT INTO users (id,email,password_hash,created_at) VALUES ($1,$2,$3,$4)`, id, c.Email, string(hash), time.Now().UTC())
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			publicError(w, 409, "registration_conflict")
			return
		}
		publicError(w, 503, "authentication_unavailable")
		return
	}
	respondToken(w, id)
}

var dummyPasswordHash, _ = bcrypt.GenerateFromPassword([]byte("timing-only-not-a-user-password"), bcrypt.DefaultCost)

func login(w http.ResponseWriter, r *http.Request) {
	c, ok := readCredentials(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	var id, hash string
	err := db.QueryRowContext(ctx, `SELECT id, password_hash FROM users WHERE email = $1`, c.Email).Scan(&id, &hash)
	if errors.Is(err, sql.ErrNoRows) {
		_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(c.Password))
		publicError(w, 401, "invalid_credentials")
		return
	}
	if err != nil {
		publicError(w, 503, "authentication_unavailable")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(c.Password)) != nil {
		publicError(w, 401, "invalid_credentials")
		return
	}
	respondToken(w, id)
}

func respondToken(w http.ResponseWriter, id string) {
	token, err := issueToken(id)
	if err != nil {
		publicError(w, 500, "authentication_failed")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, map[string]string{"token": token})
}

func issueToken(id string) (string, error) {
	if _, err := uuid.Parse(id); err != nil || len(jwtSecret) < 32 {
		return "", errors.New("invalid authentication state")
	}
	now := time.Now()
	c := jwt.RegisteredClaims{Subject: id, ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)), IssuedAt: jwt.NewNumericDate(now)}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(jwtSecret)
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := r.Header.Values("Authorization")
		if len(headers) != 1 || len(headers[0]) > 4096 {
			publicError(w, 401, "unauthenticated")
			return
		}
		parts := strings.Split(headers[0], " ")
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			publicError(w, 401, "unauthenticated")
			return
		}
		var claims jwt.RegisteredClaims
		token, err := jwt.ParseWithClaims(parts[1], &claims, func(t *jwt.Token) (any, error) { return jwtSecret, nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
		_, idErr := uuid.Parse(claims.Subject)
		if err != nil || token == nil || !token.Valid || idErr != nil || claims.IssuedAt == nil || claims.ExpiresAt == nil || !claims.ExpiresAt.After(claims.IssuedAt.Time) || len(jwtSecret) < 32 {
			publicError(w, 401, "unauthenticated")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		var exists bool
		if err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`, claims.Subject).Scan(&exists); err != nil {
			publicError(w, 503, "authentication_unavailable")
			return
		}
		if !exists {
			publicError(w, 401, "unauthenticated")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, claims.Subject)))
	})
}
