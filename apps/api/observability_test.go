package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestKaufmanRedaction(t *testing.T) {
	var out bytes.Buffer
	old := operationalLog
	operationalLog = newOperationalLogger(&out)
	defer func() { operationalLog = old }()
	ctx := context.WithValue(context.Background(), diagnosticKey{}, diagnostics{RequestID: "invalid\nsecret"})
	operationalEvent(ctx, "http.failed", slog.String("password", "secret"), slog.String("operation", "Bearer secret"), slog.Any("error", errors.New("postgres://secret")))
	var event map[string]any
	if json.Unmarshal(out.Bytes(), &event) != nil || event["event"] != "http.failed" {
		t.Fatal("invalid structured output")
	}
	if strings.Contains(out.String(), "secret") || event["request_id"] != nil {
		t.Fatal("unsafe data emitted")
	}
}

func TestQueueReceivesRequestCorrelation(t *testing.T) {
	mock, _ := securityTestSetup(t)
	t.Setenv("UPLOAD_DIR", t.TempDir())
	mock.ExpectExec("INSERT INTO batches").WillReturnResult(sqlmock.NewResult(0, 1))
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	part, _ := form.CreateFormFile("file", "fixture.zip")
	part.Write(zipFixture(t, []string{"fixture.mp4"}))
	form.Close()
	req := httptest.NewRequest("POST", "/api/batches", &body)
	req.Header.Set("Content-Type", form.FormDataContentType())
	id := "f2c726a4-46ed-4ae1-9481-b3616b66cbef"
	req.Header.Set("X-Correlation-ID", id)
	req = req.WithContext(context.WithValue(req.Context(), userKey, testUser))
	router := chi.NewRouter()
	router.Use(observationMiddleware)
	router.Post("/api/batches", createBatch)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatal(rec.Body)
	}
	payload, err := rdb.LPop(context.Background(), "jobs:batch").Result()
	if err != nil {
		t.Fatal(err)
	}
	var job map[string]string
	if json.Unmarshal([]byte(payload), &job) != nil || job["correlation_id"] != id || validID(job["request_id"]) == "" || validID(job["id"]) == "" {
		t.Fatal("queue correlation lost")
	}
}

func TestRequestCorrelationAndBoundedMetrics(t *testing.T) {
	checks := dependencyChecks{database: func(context.Context) error { return nil }, redis: func(context.Context) error { return nil }}
	router := newRouter(checks)
	req := httptest.NewRequest(http.MethodGet, "/api/health/live?password=private", nil)
	req.Header.Set("X-Request-ID", "bad-input")
	correlation := "f2c726a4-46ed-4ae1-9481-b3616b66cbef"
	req.Header.Set("X-Correlation-ID", correlation)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if validID(rec.Header().Get("X-Request-ID")) == "" || rec.Header().Get("X-Correlation-ID") != correlation {
		t.Fatal("missing correlation")
	}
	scrape := httptest.NewRecorder()
	metricsHandler(checks).ServeHTTP(scrape, httptest.NewRequest("GET", "/metrics", nil))
	if scrape.Code != 200 || !strings.Contains(scrape.Body.String(), "clipsense_http_requests_total") {
		t.Fatal("metrics missing")
	}
	if strings.Contains(scrape.Body.String(), correlation) || strings.Contains(scrape.Body.String(), "private") {
		t.Fatal("high-cardinality data in metrics")
	}
}
