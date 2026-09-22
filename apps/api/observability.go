package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type diagnosticKey struct{}
type diagnostics struct{ RequestID, CorrelationID, BatchID string }

var operationalLog = newOperationalLogger(os.Stdout)
var metricRegistry = prometheus.NewRegistry()
var events = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "clipsense_events_total", Help: "Allowlisted operational events."}, []string{"event"})
var requests = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "clipsense_http_requests_total", Help: "HTTP responses by bounded operation and status class."}, []string{"operation", "status"})
var latency = prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "clipsense_http_duration_seconds", Help: "HTTP duration.", Buckets: prometheus.DefBuckets}, []string{"operation"})
var activeUploads = prometheus.NewGauge(prometheus.GaugeOpts{Name: "clipsense_active_uploads", Help: "Active upload HTTP requests, including validation."})
var dependencyUp = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "clipsense_dependency_up", Help: "Last observed dependency availability."}, []string{"dependency"})
var queueDepth = prometheus.NewGauge(prometheus.GaugeOpts{Name: "clipsense_queue_depth", Help: "Redis pending list length; -1 means unavailable."})
var dependencyMu sync.Mutex
var dependencyState = map[string]bool{}
var eventNames = map[string]bool{}
var buildSHA = regexp.MustCompile(`^[a-f0-9]{40}$`)

func init() {
	for _, name := range []string{"service.started", "service.stopped", "service.failed", "http.completed", "http.failed", "auth.login.success", "auth.login.failed", "auth.rate_limit.triggered", "authorization.denied", "upload.started", "upload.accepted", "upload.rejected", "upload.cleanup.failed", "job.enqueued", "job.failed", "dependency.degraded", "dependency.recovered", "export.started", "export.failed", "export.completed"} {
		eventNames[name] = true
		events.WithLabelValues(name)
	}
	metricRegistry.MustRegister(events, requests, latency, activeUploads, dependencyUp, queueDepth)
}
func newOperationalLogger(w io.Writer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			a.Key = "timestamp"
		}
		if a.Key == slog.MessageKey {
			a.Key = "event"
		}
		return a
	}}))
}
func validID(s string) string {
	u, e := uuid.Parse(s)
	if e != nil || u.String() != s {
		return ""
	}
	return s
}
func diagnosticContext(ctx context.Context) diagnostics {
	d, _ := ctx.Value(diagnosticKey{}).(diagnostics)
	return d
}
func diagnosticVersion() string {
	v := os.Getenv("BUILD_SHA")
	if !buildSHA.MatchString(v) {
		return "unknown"
	}
	return v
}
func operationalEvent(ctx context.Context, name string, fields ...slog.Attr) {
	if !eventNames[name] {
		return
	}
	d := diagnosticContext(ctx)
	env := os.Getenv("APP_ENV")
	if env != "development" && env != "test" && env != "production" {
		env = "production"
	}
	attrs := []slog.Attr{slog.String("service", "api"), slog.String("environment", env), slog.String("version", diagnosticVersion())}
	for key, value := range map[string]string{"request_id": d.RequestID, "correlation_id": d.CorrelationID, "batch_id": d.BatchID} {
		if validID(value) != "" {
			attrs = append(attrs, slog.String(key, value))
		}
	}
	// Drop arbitrary fields and values: redaction is admission control, not regex scrubbing.
	for _, a := range fields {
		switch a.Key {
		case "duration_ms":
			if a.Value.Kind() == slog.KindInt64 && a.Value.Int64() >= 0 {
				attrs = append(attrs, a)
			}
		case "status":
			if a.Value.Kind() == slog.KindInt64 && a.Value.Int64() >= 100 && a.Value.Int64() <= 599 {
				attrs = append(attrs, a)
			}
		case "dependency":
			if a.Value.String() == "database" || a.Value.String() == "redis" {
				attrs = append(attrs, a)
			}
		case "operation":
			if safeOperation(a.Value.String()) != "other" {
				attrs = append(attrs, a)
			}
		case "error_code":
			switch a.Value.String() {
			case "configuration", "database", "migration", "redis", "listener", "runtime":
				attrs = append(attrs, a)
			}
		}
	}
	level := slog.LevelInfo
	if name == "service.failed" || name == "http.failed" || name == "upload.cleanup.failed" || name == "job.failed" {
		level = slog.LevelError
	}
	events.WithLabelValues(name).Inc()
	operationalLog.LogAttrs(ctx, level, name, attrs...)
}

func startupErrorCode(err error) string {
	text := err.Error()
	for prefix, code := range map[string]string{"JWT_SECRET": "configuration", "APP_ENV": "configuration", "AUTH_": "configuration", "MAX_UPLOAD": "configuration", "open db:": "database", "migrate:": "migration", "redis ping:": "redis", "listen ": "listener", "metrics listener": "listener"} {
		if strings.HasPrefix(text, prefix) {
			return code
		}
	}
	return "runtime"
}
func safeOperation(route string) string {
	switch route {
	case "/api/auth/login", "/api/auth/register", "/api/batches", "/api/batches/{id}", "/api/batches/{id}/export", "/api/health", "/api/health/live", "/api/health/ready":
		return route
	}
	return "other"
}

type observedWriter struct {
	middleware.WrapResponseWriter
	ctx context.Context
}

func (w *observedWriter) diagnosticContext() context.Context { return w.ctx }
func writerContext(w http.ResponseWriter) context.Context {
	if o, ok := w.(interface{ diagnosticContext() context.Context }); ok {
		return o.diagnosticContext()
	}
	return context.Background()
}
func observationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := validID(r.Header.Get("X-Request-ID"))
		if requestID == "" {
			requestID = uuid.NewString()
		}
		correlationID := validID(r.Header.Get("X-Correlation-ID"))
		if correlationID == "" {
			correlationID = requestID
		}
		ctx := context.WithValue(r.Context(), diagnosticKey{}, diagnostics{RequestID: requestID, CorrelationID: correlationID})
		r = r.WithContext(ctx)
		w.Header().Set("X-Request-ID", requestID)
		w.Header().Set("X-Correlation-ID", correlationID)
		ow := &observedWriter{middleware.NewWrapResponseWriter(w, r.ProtoMajor), ctx}
		start := time.Now()
		upload := r.Method == http.MethodPost && r.URL.Path == "/api/batches"
		if upload {
			activeUploads.Inc()
			defer activeUploads.Dec()
			operationalEvent(ctx, "upload.started")
		}
		defer func() {
			if recover() != nil {
				operationalEvent(ctx, "http.failed")
				http.Error(ow, "internal_error", 500)
			}
			status := ow.Status()
			if status == 0 {
				status = 200
			}
			op := safeOperation(chi.RouteContext(r.Context()).RoutePattern())
			requests.WithLabelValues(op, strconv.Itoa(status/100)+"xx").Inc()
			latency.WithLabelValues(op).Observe(time.Since(start).Seconds())
			operationalEvent(ctx, "http.completed", slog.String("operation", op), slog.Int("status", status), slog.Int64("duration_ms", time.Since(start).Milliseconds()))
			if upload && status >= 400 {
				operationalEvent(ctx, "upload.rejected", slog.Int("status", status))
			}
			if op == "/api/auth/login" || op == "/api/auth/register" {
				name := "auth.login.success"
				if status >= 400 {
					name = "auth.login.failed"
				}
				if status == 429 {
					name = "auth.rate_limit.triggered"
				}
				operationalEvent(ctx, name)
			} else if status == 401 || status == 403 {
				operationalEvent(ctx, "authorization.denied")
			}
		}()
		next.ServeHTTP(ow, r)
	})
}
func observeDependency(ctx context.Context, name string, up bool) {
	dependencyMu.Lock()
	old, known := dependencyState[name]
	dependencyState[name] = up
	dependencyMu.Unlock()
	value := 0.0
	if up {
		value = 1
	}
	dependencyUp.WithLabelValues(name).Set(value)
	if !known || old != up {
		event := "dependency.degraded"
		if up {
			event = "dependency.recovered"
		}
		operationalEvent(ctx, event, slog.String("dependency", name))
	}
}
func metricsHandler(checks dependencyChecks) http.Handler {
	handler := promhttp.HandlerFor(metricRegistry, promhttp.HandlerOpts{Timeout: 4 * time.Second, MaxRequestsInFlight: 2})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		observeDependency(ctx, "database", checks.database(ctx) == nil)
		observeDependency(ctx, "redis", checks.redis(ctx) == nil)
		depth := int64(-1)
		if rdb != nil {
			if n, e := rdb.LLen(ctx, "jobs:batch").Result(); e == nil {
				depth = n
			}
		}
		queueDepth.Set(float64(depth))
		handler.ServeHTTP(w, r)
	})
}
