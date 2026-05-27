package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
)

func TestHealthReportsDependencyFailuresAsJSON(t *testing.T) {
	originalDB := db
	originalRedis := rdb
	t.Cleanup(func() {
		db = originalDB
		rdb = originalRedis
	})

	var err error
	db, err = sql.Open("pgx", "postgres://clipsense:clipsense@127.0.0.1:1/clipsense?sslmode=disable")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	rdb = redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	t.Cleanup(func() {
		_ = rdb.Close()
	})

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()

	health(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d with body %q", http.StatusServiceUnavailable, rec.Code, rec.Body.String())
	}

	var body struct {
		Status string            `json:"status"`
		Checks map[string]string `json:"checks"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("health response was not valid JSON: %v", err)
	}
	if body.Status != "degraded" {
		t.Fatalf("expected degraded health status, got %q", body.Status)
	}
	if body.Checks["database"] != "unavailable" {
		t.Fatalf("expected database unavailable, got %q", body.Checks["database"])
	}
	if body.Checks["redis"] != "unavailable" {
		t.Fatalf("expected redis unavailable, got %q", body.Checks["redis"])
	}
}

func TestLocalDevCORSAllowsConfiguredOriginPreflight(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/api/batches", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	localDevCORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not run for preflight")
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("expected localhost origin, got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(got, "Authorization") {
		t.Fatalf("expected Authorization in allowed headers, got %q", got)
	}
}

func TestLocalDevCORSRejectsUnexpectedOriginPreflight(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/api/batches", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()

	localDevCORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not run for rejected preflight")
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestAuthMiddlewareRejectsMissingToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/batches", nil)
	rec := httptest.NewRecorder()

	authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not run without token")
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthMiddlewareRejectsInvalidToken(t *testing.T) {
	jwtSecret = []byte("test-secret")
	req := httptest.NewRequest(http.MethodGet, "/api/batches", nil)
	req.Header.Set("Authorization", "Bearer not-a-valid-token")
	rec := httptest.NewRecorder()

	authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not run with invalid token")
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestCreateBatchRejectsMissingFileBeforePersistence(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("name", "missing-file"); err != nil {
		t.Fatalf("write multipart field: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/batches", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	createBatch(rec, req)

	if rec.Code < 400 {
		t.Fatalf("expected batch creation without file to fail, got status %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "missing file") {
		t.Fatalf("expected missing file error, got %q", rec.Body.String())
	}
}
