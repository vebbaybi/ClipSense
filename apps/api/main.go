package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
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
		operationalEvent(context.Background(), "service.failed", slog.String("error_code", startupErrorCode(err)))
		os.Exit(1)
	}
}

func run() error {
	var err error
	security, err = loadSecurityConfig()
	if err != nil {
		return err
	}
	jwtSecret = security.secret
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
	metricsListener, err := net.Listen("tcp", getenv("METRICS_ADDR", "127.0.0.1:9091"))
	if err != nil {
		listener.Close()
		return errors.New("metrics listener unavailable")
	}
	metricsServer := &http.Server{Handler: metricsHandler(runtimeDependencyChecks()), ReadHeaderTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second}
	defer metricsServer.Close()
	go func() {
		if e := metricsServer.Serve(metricsListener); e != nil && !errors.Is(e, http.ErrServerClosed) {
			operationalEvent(context.Background(), "service.failed")
			stop()
		}
	}()
	operationalEvent(context.Background(), "service.started")
	return serveHTTP(shutdownCtx, server, listener, serverShutdownTimeout)
}

func newRouter(checks dependencyChecks) http.Handler {
	r := chi.NewRouter()
	r.Use(observationMiddleware)
	r.Use(localDevCORSMiddleware)

	r.Route("/api", func(api chi.Router) {
		api.Get("/health", readinessHandler(checks))
		api.Get("/health/live", liveness)
		api.Get("/health/ready", readinessHandler(checks))
		api.With(authAttemptLimit).Post("/auth/register", register)
		api.With(authAttemptLimit).Post("/auth/login", login)
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
	operationalEvent(context.Background(), "service.stopped")
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
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID, X-Correlation-ID")
			w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID, X-Correlation-ID")
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
		}
		if err := dependencies.redis(ctx); err != nil {
			checks["redis"] = "unavailable"
			statusCode = http.StatusServiceUnavailable
		}

		observeDependency(r.Context(), "database", checks["database"] == "ok")
		observeDependency(r.Context(), "redis", checks["redis"] == "ok")
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
             FROM storyline_clips JOIN clips ON clips.id = storyline_clips.clip_id WHERE storyline_clips.storyline_id = $1 AND clips.batch_id = $2 ORDER BY storyline_clips.position`, s.ID, id)
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
	operationalEvent(r.Context(), "export.started")
	started := time.Now()
	success := false
	defer func() {
		event := "export.failed"
		if success {
			event = "export.completed"
		}
		operationalEvent(r.Context(), event, slog.Int64("duration_ms", time.Since(started).Milliseconds()))
	}()
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
		success = cw.Error() == nil
		return
	}
	writeJSON(w, map[string]any{"batch": b, "clips": clips, "storylines": storylines})
	success = true
}

func httpError(w http.ResponseWriter, err error) {
	operationalEvent(writerContext(w), "http.failed")
	http.Error(w, "internal_error", http.StatusInternalServerError)
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
func redisOpts(url string) *redis.Options {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return &redis.Options{Addr: "redis:6379"}
	}
	return opt
}
