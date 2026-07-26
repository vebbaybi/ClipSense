package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLivenessReportsProcessAlive(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/health/live", nil)
	rec := httptest.NewRecorder()

	liveness(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %q", http.StatusOK, rec.Code, rec.Body.String())
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Status != "ok" {
		t.Fatalf("unexpected liveness response: %q (%v)", rec.Body.String(), err)
	}
}

func TestReadinessReportsSuccess(t *testing.T) {
	rec := runReadinessTest(t, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %q", http.StatusOK, rec.Code, rec.Body.String())
	}

	var body struct {
		Status string            `json:"status"`
		Checks map[string]string `json:"checks"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("readiness response was not valid JSON: %v", err)
	}
	if body.Status != "ok" || body.Checks["database"] != "ok" || body.Checks["redis"] != "ok" {
		t.Fatalf("unexpected readiness response: %q", rec.Body.String())
	}
}

func TestReadinessReportsDatabaseFailureSafely(t *testing.T) {
	rec := runReadinessTest(t, errors.New("private database detail"), nil)
	assertReadinessFailure(t, rec, "database", "private database detail")
}

func TestReadinessReportsRedisFailureSafely(t *testing.T) {
	rec := runReadinessTest(t, nil, errors.New("private redis detail"))
	assertReadinessFailure(t, rec, "redis", "private redis detail")
}

func TestHealthAliasUsesReadiness(t *testing.T) {
	checks := dependencyChecks{
		database: func(context.Context) error { return nil },
		redis:    func(context.Context) error { return nil },
	}
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()

	newRouter(checks).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"database":"ok"`) {
		t.Fatalf("health alias did not return readiness: %d %q", rec.Code, rec.Body.String())
	}
}

func TestServeHTTPShutsDownAfterCancellation(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	server := &http.Server{Handler: http.HandlerFunc(liveness)}
	done := make(chan error, 1)
	go func() {
		done <- serveHTTP(ctx, server, listener, time.Second)
	}()

	response, err := http.Get("http://" + listener.Addr().String())
	if err != nil {
		cancel()
		t.Fatalf("request server: %v", err)
	}
	_ = response.Body.Close()
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("graceful shutdown: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not shut down within deadline")
	}
}

func runReadinessTest(t *testing.T, databaseErr, redisErr error) *httptest.ResponseRecorder {
	t.Helper()
	checks := dependencyChecks{
		database: func(context.Context) error { return databaseErr },
		redis:    func(context.Context) error { return redisErr },
	}
	req := httptest.NewRequest(http.MethodGet, "/api/health/ready", nil)
	rec := httptest.NewRecorder()
	readinessHandler(checks).ServeHTTP(rec, req)
	return rec
}

func assertReadinessFailure(t *testing.T, rec *httptest.ResponseRecorder, failedCheck, privateDetail string) {
	t.Helper()
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d with body %q", http.StatusServiceUnavailable, rec.Code, rec.Body.String())
	}

	var body struct {
		Status string            `json:"status"`
		Checks map[string]string `json:"checks"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("readiness response was not valid JSON: %v", err)
	}
	if body.Status != "degraded" || body.Checks[failedCheck] != "unavailable" {
		t.Fatalf("unexpected degraded readiness response: %q", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), privateDetail) {
		t.Fatalf("readiness response leaked dependency error: %q", rec.Body.String())
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
