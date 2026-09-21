package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB
var rdb *redis.Client
var jwtSecret []byte

const (
	userKey                 = "user_id"
	healthDependencyTimeout = 2 * time.Second
	serverReadHeaderTimeout = 5 * time.Second
	serverReadTimeout       = 5 * time.Minute
	serverWriteTimeout      = 5 * time.Minute
	serverIdleTimeout       = 1 * time.Minute
	serverShutdownTimeout   = 15 * time.Second
)

type dependencyChecks struct {
	database func(context.Context) error
	redis    func(context.Context) error
}

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type Batch struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	Name            string    `json:"name"`
	Status          string    `json:"status"`
	ZipPath         string    `json:"zip_path"`
	ClipCount       int       `json:"clip_count"`
	DurationSeconds float64   `json:"duration_seconds"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Clip struct {
	ID         string    `json:"id"`
	BatchID    string    `json:"batch_id"`
	Filename   string    `json:"filename"`
	Title      string    `json:"title"`
	Summary    string    `json:"summary"`
	Transcript string    `json:"transcript"`
	Mood       string    `json:"mood"`
	Role       string    `json:"role"`
	Topic      string    `json:"topic"`
	Duration   float64   `json:"duration_seconds"`
	CreatedAt  time.Time `json:"created_at"`
}

type Storyline struct {
	ID        string    `json:"id"`
	BatchID   string    `json:"batch_id"`
	Title     string    `json:"title"`
	Type      string    `json:"type"`
	Clips     []Clip    `json:"clips"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	jwtSecret = []byte(getenv("JWT_SECRET", "dev-secret"))

	var err error
	db, err = openDB()
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()
	if err := migrate(); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	redisURL := getenv("REDIS_URL", "redis://redis:6379")
	rdb = redis.NewClient(redisOpts(redisURL))
	defer rdb.Close()
	redisCtx, redisCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer redisCancel()
	if err := rdb.Ping(redisCtx).Err(); err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}

	server := &http.Server{
		Addr:              getenv("API_ADDR", ":8080"),
		Handler:           newRouter(runtimeDependencyChecks()),
		ReadHeaderTimeout: serverReadHeaderTimeout,
		ReadTimeout:       serverReadTimeout,
		WriteTimeout:      serverWriteTimeout,
		IdleTimeout:       serverIdleTimeout,
	}
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", server.Addr, err)
	}

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	log.Printf("api listening on %s", server.Addr)
	return serveHTTP(shutdownCtx, server, listener, serverShutdownTimeout)
}

func newRouter(checks dependencyChecks) http.Handler {
	r := chi.NewRouter()
	r.Use(localDevCORSMiddleware)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api", func(api chi.Router) {
		api.Get("/health", readinessHandler(checks))
		api.Get("/health/live", liveness)
		api.Get("/health/ready", readinessHandler(checks))
		api.Post("/auth/register", register)
		api.Post("/auth/login", login)
		api.Group(func(p chi.Router) {
			p.Use(authMiddleware)
			p.Get("/batches", listBatches)
			p.Post("/batches", createBatch)
			p.Get("/batches/{id}", getBatch)
			p.Get("/batches/{id}/export", exportBatch)
		})
	})

	return r
}

func runtimeDependencyChecks() dependencyChecks {
	return dependencyChecks{
		database: func(ctx context.Context) error {
			return db.PingContext(ctx)
		},
		redis: func(ctx context.Context) error {
			return rdb.Ping(ctx).Err()
		},
	}
}

func serveHTTP(ctx context.Context, server *http.Server, listener net.Listener, shutdownTimeout time.Duration) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve(listener)
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	case <-ctx.Done():
		log.Printf("api shutdown started")
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(timeoutCtx); err != nil {
		_ = server.Close()
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}

	if err := <-errCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP during shutdown: %w", err)
	}
	log.Printf("api shutdown complete")
	return nil
}

func localDevCORSMiddleware(next http.Handler) http.Handler {
	allowedOrigins := map[string]bool{
		"http://localhost:3000": true,
		"http://127.0.0.1:3000": true,
	}
	if configured := os.Getenv("CORS_ALLOWED_ORIGINS"); configured != "" {
		allowedOrigins = map[string]bool{}
		for _, origin := range strings.Split(configured, ",") {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				allowedOrigins[origin] = true
			}
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Vary", "Origin")
		}

		if r.Method == http.MethodOptions {
			if origin != "" && !allowedOrigins[origin] {
				http.Error(w, "cors origin not allowed", http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func liveness(w http.ResponseWriter, _ *http.Request) {
	writeJSONStatus(w, http.StatusOK, map[string]string{"status": "ok"})
}

func readinessHandler(dependencies dependencyChecks) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		statusCode := http.StatusOK
		checks := map[string]string{
			"database": "ok",
			"redis":    "ok",
		}

		ctx, cancel := context.WithTimeout(r.Context(), healthDependencyTimeout)
		defer cancel()

		if err := dependencies.database(ctx); err != nil {
			checks["database"] = "unavailable"
			statusCode = http.StatusServiceUnavailable
			log.Printf("readiness database unavailable: %v", err)
		}
		if err := dependencies.redis(ctx); err != nil {
			checks["redis"] = "unavailable"
			statusCode = http.StatusServiceUnavailable
			log.Printf("readiness redis unavailable: %v", err)
		}

		status := "ok"
		if statusCode != http.StatusOK {
			status = "degraded"
		}

		writeJSONStatus(w, statusCode, map[string]any{
			"status": status,
			"checks": checks,
		})
	}
}

// Auth
func register(w http.ResponseWriter, r *http.Request) {
	var body struct{ Email, Password string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpError(w, err)
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	id := uuid.New().String()
	now := time.Now().UTC()
	_, err := db.Exec(`INSERT INTO users (id,email,password_hash,created_at) VALUES ($1,$2,$3,$4)`, id, body.Email, string(hash), now)
	if err != nil {
		httpError(w, err)
		return
	}
	token, _ := issueToken(id)
	writeJSON(w, map[string]string{"token": token})
}

func login(w http.ResponseWriter, r *http.Request) {
	var body struct{ Email, Password string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpError(w, err)
		return
	}
	var id, hash string
	err := db.QueryRow(`SELECT id, password_hash FROM users WHERE email = $1`, body.Email).Scan(&id, &hash)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(body.Password)) != nil {
		httpError(w, fmt.Errorf("invalid credentials"))
		return
	}
	token, _ := issueToken(id)
	writeJSON(w, map[string]string{"token": token})
}

func issueToken(userID string) (string, error) {
	claims := jwt.MapClaims{"sub": userID, "exp": time.Now().Add(24 * time.Hour).Unix(), "iat": time.Now().Unix()}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "missing auth", http.StatusUnauthorized)
			return
		}
		var tokenStr string
		fmt.Sscanf(auth, "Bearer %s", &tokenStr)
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("invalid method")
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		claims := token.Claims.(jwt.MapClaims)
		uid, _ := claims["sub"].(string)
		ctx := context.WithValue(r.Context(), userKey, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func currentUserID(r *http.Request) string {
	if v := r.Context().Value(userKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// DB
func openDB() (*sql.DB, error) {
	dsn := getenv("DATABASE_URL", "postgres://clipsense:clipsense@postgres:5432/clipsense?sslmode=disable")
	return sql.Open("pgx", dsn)
}

func migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (id TEXT PRIMARY KEY, email TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL, created_at TIMESTAMPTZ);`,
		`CREATE TABLE IF NOT EXISTS batches (id TEXT PRIMARY KEY, user_id TEXT, name TEXT, status TEXT, zip_path TEXT, clip_count INTEGER DEFAULT 0, duration_seconds REAL DEFAULT 0, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ);`,
		`CREATE TABLE IF NOT EXISTS clips (id TEXT PRIMARY KEY, batch_id TEXT, filename TEXT, title TEXT, summary TEXT, transcript TEXT, mood TEXT, role TEXT, topic TEXT, duration_seconds REAL, created_at TIMESTAMPTZ);`,
		`CREATE TABLE IF NOT EXISTS storylines (id TEXT PRIMARY KEY, batch_id TEXT, title TEXT, type TEXT, created_at TIMESTAMPTZ);`,
		`CREATE TABLE IF NOT EXISTS storyline_clips (storyline_id TEXT, clip_id TEXT, position INTEGER, PRIMARY KEY(storyline_id, position));`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	return nil
}

// loadBatch retrieves batch, clips, and storylines for a user.
func loadBatch(uid, id string) (Batch, []Clip, []Storyline, error) {
	var b Batch
	err := db.QueryRow(`SELECT id, user_id, name, status, zip_path, clip_count, duration_seconds, created_at, updated_at FROM batches WHERE id = $1 AND user_id = $2`, id, uid).
		Scan(&b.ID, &b.UserID, &b.Name, &b.Status, &b.ZipPath, &b.ClipCount, &b.DurationSeconds, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return b, nil, nil, err
	}
	clipsRows, err := db.Query(`SELECT id, batch_id, filename, title, summary, transcript, mood, role, topic, duration_seconds, created_at FROM clips WHERE batch_id = $1 ORDER BY created_at`, id)
	if err != nil {
		return b, nil, nil, err
	}
	var clips []Clip
	for clipsRows.Next() {
		var c Clip
		if err := clipsRows.Scan(&c.ID, &c.BatchID, &c.Filename, &c.Title, &c.Summary, &c.Transcript, &c.Mood, &c.Role, &c.Topic, &c.Duration, &c.CreatedAt); err != nil {
			return b, nil, nil, err
		}
		clips = append(clips, c)
	}
	sRows, err := db.Query(`SELECT id, batch_id, title, type, created_at FROM storylines WHERE batch_id = $1`, id)
	if err != nil {
		return b, nil, nil, err
	}
	var storylines []Storyline
	for sRows.Next() {
		var s Storyline
		if err := sRows.Scan(&s.ID, &s.BatchID, &s.Title, &s.Type, &s.CreatedAt); err != nil {
			return b, nil, nil, err
		}
		crows, err := db.Query(`SELECT clips.id, clips.batch_id, clips.filename, clips.title, clips.summary, clips.transcript, clips.mood, clips.role, clips.topic, clips.duration_seconds, clips.created_at
             FROM storyline_clips JOIN clips ON clips.id = storyline_clips.clip_id WHERE storyline_clips.storyline_id = $1 ORDER BY storyline_clips.position`, s.ID)
		if err != nil {
			return b, nil, nil, err
		}
		for crows.Next() {
			var c Clip
			if err := crows.Scan(&c.ID, &c.BatchID, &c.Filename, &c.Title, &c.Summary, &c.Transcript, &c.Mood, &c.Role, &c.Topic, &c.Duration, &c.CreatedAt); err != nil {
				return b, nil, nil, err
			}
			s.Clips = append(s.Clips, c)
		}
		storylines = append(storylines, s)
	}
	return b, clips, storylines, nil
}

// Batches
func listBatches(w http.ResponseWriter, r *http.Request) {
	uid := currentUserID(r)
	rows, err := db.Query(`SELECT id, user_id, name, status, clip_count, duration_seconds, created_at, updated_at FROM batches WHERE user_id = $1 ORDER BY created_at DESC`, uid)
	if err != nil {
		httpError(w, err)
		return
	}
	var batches []Batch
	for rows.Next() {
		var b Batch
		if err := rows.Scan(&b.ID, &b.UserID, &b.Name, &b.Status, &b.ClipCount, &b.DurationSeconds, &b.CreatedAt, &b.UpdatedAt); err != nil {
			httpError(w, err)
			return
		}
		batches = append(batches, b)
	}
	writeJSON(w, map[string]any{"batches": batches})
}

func createBatch(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(200 << 20); err != nil {
		httpError(w, err)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		httpError(w, fmt.Errorf("missing file: %w", err))
		return
	}
	defer file.Close()
	name := r.FormValue("name")
	if name == "" {
		name = header.Filename
	}
	id := uuid.New().String()
	now := time.Now().UTC()
	uid := currentUserID(r)

	uploads := getenv("UPLOAD_DIR", "data/uploads")
	if err := os.MkdirAll(uploads, 0o755); err != nil {
		httpError(w, err)
		return
	}
	destPath := filepath.Join(uploads, fmt.Sprintf("%s.zip", id))
	out, err := os.Create(destPath)
	if err != nil {
		httpError(w, err)
		return
	}
	if _, err := ioCopy(out, file); err != nil {
		httpError(w, err)
		return
	}
	out.Close()

	_, err = db.Exec(`INSERT INTO batches (id, user_id, name, status, zip_path, clip_count, duration_seconds, created_at, updated_at) VALUES ($1,$2,$3,'pending',$4,0,0,$5,$6)`,
		id, uid, name, destPath, now, now)
	if err != nil {
		httpError(w, err)
		return
	}

	job := map[string]string{"id": id, "name": name, "zip_path": destPath}
	payload, err := json.Marshal(job)
	if err != nil {
		httpError(w, err)
		return
	}
	enqueueCtx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	if err := rdb.RPush(enqueueCtx, "jobs:batch", payload).Err(); err != nil {
		_, _ = db.Exec(`UPDATE batches SET status='failed', updated_at=CURRENT_TIMESTAMP WHERE id=$1`, id)
		httpError(w, fmt.Errorf("enqueue batch job: %w", err))
		return
	}

	writeJSON(w, map[string]any{"batch": Batch{ID: id, UserID: uid, Name: name, Status: "pending", ZipPath: destPath, CreatedAt: now, UpdatedAt: now}})
}

func getBatch(w http.ResponseWriter, r *http.Request) {
	uid := currentUserID(r)
	id := chi.URLParam(r, "id")
	b, clips, storylines, err := loadBatch(uid, id)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		httpError(w, err)
		return
	}
	writeJSON(w, map[string]any{"batch": b, "clips": clips, "storylines": storylines})
}

func exportBatch(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}
	id := chi.URLParam(r, "id")
	uid := currentUserID(r)
	b, clips, storylines, err := loadBatch(uid, id)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		httpError(w, err)
		return
	}
	if format == "csv" {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=batch_%s.csv", id))
		cw := csv.NewWriter(w)
		cw.Write([]string{"clip_id", "title", "summary", "mood", "role", "duration_seconds"})
		for _, c := range clips {
			cw.Write([]string{c.ID, c.Title, c.Summary, c.Mood, c.Role, fmt.Sprint(c.Duration)})
		}
		cw.Flush()
		return
	}
	writeJSON(w, map[string]any{"batch": b, "clips": clips, "storylines": storylines})
}

func httpError(w http.ResponseWriter, err error) {
	log.Println("error", err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
func writeJSON(w http.ResponseWriter, payload any) { writeJSONStatus(w, http.StatusOK, payload) }
func writeJSONStatus(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
func ioCopy(dst *os.File, src multipart.File) (int64, error) { return io.Copy(dst, src) }
func redisOpts(url string) *redis.Options {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return &redis.Options{Addr: "redis:6379"}
	}
	return opt
}
