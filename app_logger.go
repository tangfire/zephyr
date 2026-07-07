package main

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const requestIDHeader = "X-Request-ID"

type statusResponseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *statusResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusResponseWriter) Write(payload []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(payload)
	w.bytes += n
	return n, err
}

func initAppLogger(cfg Config) (*zap.Logger, func(), error) {
	encoder := zap.NewProductionEncoderConfig()
	encoder.TimeKey = "ts"
	encoder.LevelKey = "level"
	encoder.NameKey = "logger"
	encoder.CallerKey = "caller"
	encoder.MessageKey = "message"
	encoder.StacktraceKey = "stacktrace"
	encoder.EncodeLevel = zapcore.LowercaseLevelEncoder
	encoder.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder.EncodeDuration = zapcore.MillisDurationEncoder

	level := zap.NewAtomicLevelAt(zap.InfoLevel)
	if parsed, err := zapcore.ParseLevel(strings.ToLower(strings.TrimSpace(cfg.LogLevel))); err == nil {
		level.SetLevel(parsed)
	}
	zapConfig := zap.Config{
		Level:            level,
		Development:      !strings.EqualFold(cfg.AppEnv, "production"),
		Encoding:         "json",
		EncoderConfig:    encoder,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields: map[string]any{
			"service": "peapod",
			"env":     cfg.AppEnv,
		},
	}
	logger, err := zapConfig.Build(zap.AddCaller())
	if err != nil {
		return nil, func() {}, err
	}
	undoGlobal := zap.ReplaceGlobals(logger)
	undoStdLog := zap.RedirectStdLog(logger.Named("stdlog"))
	log.SetFlags(0)
	cleanup := func() {
		undoStdLog()
		undoGlobal()
		syncLogger(logger)
	}
	return logger, cleanup, nil
}

func accessLogMiddleware(logger *zap.Logger, cfg Config, next http.Handler) http.Handler {
	mode := normalizeAccessLogMode(cfg.AccessLogMode)
	slowThreshold := time.Duration(cfg.AccessLogSlowThresholdSeconds) * time.Second
	if slowThreshold <= 0 {
		slowThreshold = 3 * time.Second
	}
	if logger == nil {
		logger = zap.L()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := normalizeRequestID(r.Header.Get(requestIDHeader))
		if requestID == "" {
			requestID = newRequestID()
		}
		w.Header().Set(requestIDHeader, requestID)
		r.Header.Set(requestIDHeader, requestID)

		start := time.Now()
		rw := &statusResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		latency := time.Since(start)
		status := rw.status
		if !shouldWriteHTTPAccessLog(mode, status, latency, slowThreshold, requestPath(r)) {
			return
		}
		fields := []zap.Field{
			zap.String("event", "http_request"),
			zap.String("method", r.Method),
			zap.String("path", requestPath(r)),
			zap.Int("status", status),
			zap.String("status_class", httpStatusClass(status)),
			zap.Int64("latency_ms", latency.Milliseconds()),
			zap.Int("response_size_bytes", rw.bytes),
			zap.String("client_ip", requestClientIP(r)),
			zap.String("request_id", requestID),
		}
		switch {
		case status >= 500:
			logger.Error("HTTP request failed", fields...)
		case status >= 400 || latency >= slowThreshold:
			logger.Warn("HTTP request needs attention", fields...)
		default:
			logger.Info("HTTP request completed", fields...)
		}
	})
}

func normalizeAccessLogMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "all", "full":
		return "all"
	case "off", "none", "disabled":
		return "off"
	default:
		return "attention"
	}
}

func shouldWriteHTTPAccessLog(mode string, status int, latency time.Duration, slowThreshold time.Duration, path string) bool {
	switch normalizeAccessLogMode(mode) {
	case "all":
		if isHealthPath(path) && status < 400 {
			return false
		}
		return true
	case "off":
		return false
	default:
		return status >= 400 || latency >= slowThreshold
	}
}

func isHealthPath(path string) bool {
	switch strings.TrimSpace(path) {
	case "/health", "/healthz", "/ready", "/readyz", "/api/health":
		return true
	default:
		return false
	}
}

func httpStatusClass(status int) string {
	switch {
	case status >= 100 && status < 200:
		return "1xx"
	case status >= 200 && status < 300:
		return "2xx"
	case status >= 300 && status < 400:
		return "3xx"
	case status >= 400 && status < 500:
		return "4xx"
	case status >= 500 && status < 600:
		return "5xx"
	default:
		return "unknown"
	}
}

func requestPath(r *http.Request) string {
	if r == nil || r.URL == nil {
		return ""
	}
	return r.URL.Path
}

func requestClientIP(r *http.Request) string {
	for _, header := range []string{"X-Real-IP", "X-Forwarded-For"} {
		value := strings.TrimSpace(r.Header.Get(header))
		if value == "" {
			continue
		}
		if header == "X-Forwarded-For" {
			value = strings.TrimSpace(strings.Split(value, ",")[0])
		}
		if value != "" {
			return value
		}
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func normalizeRequestID(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 96 {
		value = value[:96]
	}
	var b strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func newRequestID() string {
	var b [16]byte
	if _, err := io.ReadFull(rand.Reader, b[:]); err != nil {
		return "req_local"
	}
	return "req_" + hex.EncodeToString(b[:])
}

func syncLogger(logger *zap.Logger) {
	if logger == nil {
		return
	}
	if err := logger.Sync(); err != nil && !strings.Contains(strings.ToLower(err.Error()), "invalid argument") {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
	}
}
