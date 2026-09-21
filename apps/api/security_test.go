package main

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

const testUser = "9ba159c9-4d80-4d28-b64e-062842118f67"

func securityTestSetup(t *testing.T) (sqlmock.Sqlmock, *miniredis.Miniredis) {
	t.Helper()
	oldDB, oldRedis, oldSecret, oldConfig := db, rdb, jwtSecret, security
	var mock sqlmock.Sqlmock
	var err error
	db, mock, err = sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	server := miniredis.RunT(t)
	rdb = redis.NewClient(&redis.Options{Addr: server.Addr(), MaxRetries: -1})
	jwtSecret = []byte("deterministic-test-signing-material-0123456789")
	security = securityConfig{secret: jwtSecret, authAttempts: 2, authWindow: time.Minute, uploadBytes: 1 << 20}
	t.Cleanup(func() {
		db.Close()
		rdb.Close()
		db, rdb, jwtSecret, security = oldDB, oldRedis, oldSecret, oldConfig
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})
	return mock, server
}

func TestSecretConfiguration(t *testing.T) {
	for _, tc := range []struct {
		mode, secret string
		valid        bool
	}{
		{"", "", false}, {"production", "dev-secret", false}, {"production", "not-base64", false},
		{"test", "", false}, {"development", "", true}, {"developmnt", "", false},
		{"production", base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0}, 32)), false},
		{"production", base64.StdEncoding.EncodeToString([]byte("deterministic-test-material-123456")), true},
	} {
		t.Run(tc.mode+tc.secret, func(t *testing.T) {
			t.Setenv("APP_ENV", tc.mode)
			t.Setenv("JWT_SECRET", tc.secret)
			_, err := loadSecurityConfig()
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v error=%v", tc.valid, err)
			}
		})
	}
}

func TestUnsignedDuplicateAndBackendAuthFailures(t *testing.T) {
	mock, _ := securityTestSetup(t)
	good, err := issueToken(testUser)
	if err != nil {
		t.Fatal(err)
	}
	unsigned := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.RegisteredClaims{Subject: testUser, ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))})
	none, _ := unsigned.SignedString(jwt.UnsafeAllowNoneSignatureType)
	for _, header := range []string{none, strings.Repeat("x", 5000)} {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+header)
		rec := httptest.NewRecorder()
		authMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("bad token accepted") })).ServeHTTP(rec, req)
		if rec.Code != 401 {
			t.Fatal(rec.Code)
		}
	}
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add("Authorization", "Bearer "+good)
	req.Header.Add("Authorization", "Bearer "+good)
	rec := httptest.NewRecorder()
	authMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("duplicate header accepted") })).ServeHTTP(rec, req)
	if rec.Code != 401 {
		t.Fatal(rec.Code)
	}
	mock.ExpectQuery("SELECT EXISTS").WillReturnError(errors.New("private SQL host path"))
	req.Header.Set("Authorization", "Bearer "+good)
	rec = httptest.NewRecorder()
	authMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("backend outage accepted") })).ServeHTTP(rec, req)
	if rec.Code != 503 || strings.Contains(rec.Body.String(), "private") {
		t.Fatal(rec.Body)
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	mock.ExpectQuery("SELECT id").WillReturnRows(sqlmock.NewRows([]string{"id", "password_hash"}).AddRow(testUser, string(hash)))
	rec = httptest.NewRecorder()
	login(rec, httptest.NewRequest("POST", "/", strings.NewReader(`{"email":"x@y.test","password":"wrong-password"}`)))
	if rec.Code != 401 {
		t.Fatal(rec.Body)
	}
}

func TestAuthenticationBoundary(t *testing.T) {
	mock, _ := securityTestSetup(t)
	now := time.Now()
	valid := jwt.RegisteredClaims{Subject: testUser, ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)), IssuedAt: jwt.NewNumericDate(now)}
	sign := func(c any, method jwt.SigningMethod, key []byte) string {
		var claims jwt.Claims
		switch v := c.(type) {
		case jwt.RegisteredClaims:
			claims = v
		case jwt.MapClaims:
			claims = v
		}
		token, err := jwt.NewWithClaims(method, claims).SignedString(key)
		if err != nil {
			t.Fatal(err)
		}
		return token
	}
	good := sign(valid, jwt.SigningMethodHS256, jwtSecret)
	expired := valid
	expired.ExpiresAt = jwt.NewNumericDate(now.Add(-time.Hour))
	noID := valid
	noID.Subject = ""
	noExpiry := valid
	noExpiry.ExpiresAt = nil
	noIssued := valid
	noIssued.IssuedAt = nil
	future := valid
	future.IssuedAt = jwt.NewNumericDate(now.Add(time.Hour))
	for _, tc := range []struct {
		name, header string
		code         int
	}{
		{"missing", "", 401}, {"scheme", "Basic " + good, 401}, {"extra", "Bearer " + good + " junk", 401},
		{"empty", "Bearer ", 401}, {"malformed", "Bearer private.raw.token", 401},
		{"signature", "Bearer " + good[:len(good)-8] + "AAAAAAAA", 401},
		{"wrong-secret", "Bearer " + sign(valid, jwt.SigningMethodHS256, []byte("wrong-secret")), 401},
		{"expired", "Bearer " + sign(expired, jwt.SigningMethodHS256, jwtSecret), 401},
		{"algorithm", "Bearer " + sign(valid, jwt.SigningMethodHS384, jwtSecret), 401},
		{"identity", "Bearer " + sign(noID, jwt.SigningMethodHS256, jwtSecret), 401},
		{"expiry-required", "Bearer " + sign(noExpiry, jwt.SigningMethodHS256, jwtSecret), 401},
		{"issued-required", "Bearer " + sign(noIssued, jwt.SigningMethodHS256, jwtSecret), 401},
		{"future-issued", "Bearer " + sign(future, jwt.SigningMethodHS256, jwtSecret), 401},
		{"claim-type", "Bearer " + sign(jwt.MapClaims{"sub": 42, "exp": "tomorrow"}, jwt.SigningMethodHS256, jwtSecret), 401},
		{"valid", "Bearer " + good, 204},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.code == 204 {
				mock.ExpectQuery("SELECT EXISTS").WithArgs(testUser).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
			}
			req := httptest.NewRequest("GET", "/api/batches", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			rec := httptest.NewRecorder()
			authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if currentUserID(r) != testUser {
					t.Error("wrong identity")
				}
				w.WriteHeader(204)
			})).ServeHTTP(rec, req)
			if rec.Code != tc.code {
				t.Fatalf("got %d %s", rec.Code, rec.Body)
			}
			if strings.Contains(rec.Body.String(), good) || strings.Contains(rec.Body.String(), string(jwtSecret)) || strings.Contains(rec.Body.String(), "private.raw.token") {
				t.Fatal("leaked auth material")
			}
		})
	}
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+good)
	rec := httptest.NewRecorder()
	authMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("deleted user accepted") })).ServeHTTP(rec, req)
	if rec.Code != 401 {
		t.Fatal(rec.Code)
	}
}

func TestAttemptLimitExpiryAndOutage(t *testing.T) {
	_, server := securityTestSetup(t)
	handler := authAttemptLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	request := func(ip string) int {
		req := httptest.NewRequest("POST", "/api/auth/login", nil)
		req.RemoteAddr = ip
		req.Header.Set("X-Forwarded-For", "1.2.3.4")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec.Code
	}
	for _, want := range []int{204, 204, 429} {
		if got := request("192.0.2.1:123"); got != want {
			t.Fatalf("got %d want %d", got, want)
		}
	}
	if got := request("192.0.2.2:123"); got != 204 {
		t.Fatal("unrelated peer locked out")
	}
	server.FastForward(time.Minute + time.Second)
	if got := request("192.0.2.1:123"); got != 204 {
		t.Fatal("expiry did not recover")
	}
	server.SetError("private backend failure")
	if got := request("192.0.2.1:123"); got != 503 {
		t.Fatal("limiter failed open")
	}
	server.SetError("")
	if got := request("192.0.2.1:123"); got != 204 {
		t.Fatal("limiter did not recover")
	}
}

func TestCredentialValidationAndHandlers(t *testing.T) {
	mock, _ := securityTestSetup(t)
	for _, body := range []string{`{}`, `{"email":"bad","password":"12345678"}`, `{"email":"x@y.test","password":""}`, `{"email":"x@y.test","password":"` + strings.Repeat("x", 73) + `"}`, `{"email":"x@y.test","password":"12345678"} {}`, strings.Repeat("x", 5000)} {
		rec := httptest.NewRecorder()
		register(rec, httptest.NewRequest("POST", "/", strings.NewReader(body)))
		if rec.Code != 400 {
			t.Fatalf("invalid credentials status %d", rec.Code)
		}
	}
	body := `{"email":"x@y.test","password":"test-password"}`
	mock.ExpectExec("INSERT INTO users").WillReturnResult(sqlmock.NewResult(0, 1))
	rec := httptest.NewRecorder()
	register(rec, httptest.NewRequest("POST", "/", strings.NewReader(body)))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "token") {
		t.Fatal(rec.Body)
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte("test-password"), bcrypt.MinCost)
	mock.ExpectQuery("SELECT id").WillReturnRows(sqlmock.NewRows([]string{"id", "password_hash"}).AddRow(testUser, string(hash)))
	rec = httptest.NewRecorder()
	login(rec, httptest.NewRequest("POST", "/", strings.NewReader(body)))
	if rec.Code != 200 {
		t.Fatal(rec.Body)
	}
	for _, err := range []error{sql.ErrNoRows, errors.New("secret SQL filesystem detail")} {
		mock.ExpectQuery("SELECT id").WillReturnError(err)
		rec = httptest.NewRecorder()
		login(rec, httptest.NewRequest("POST", "/", strings.NewReader(body)))
		if rec.Code < 400 || strings.Contains(rec.Body.String(), "SQL") {
			t.Fatal(rec.Body)
		}
	}
}
