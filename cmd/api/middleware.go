package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"task-manager-api/internal/httputil"
	"time"
)

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

type contextKey string

const requestIDKey contextKey = "request_id"

func (w *statusResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()

		wrappedWriter := &statusResponseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}
		next.ServeHTTP(wrappedWriter, r)

		requestID, _ := r.Context().Value(requestIDKey).(string)
		logger.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrappedWriter.status,
			"duration", time.Since(startedAt),
			"request_id", requestID,
		)
	})
}

func requestIDMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID, err := generateRequestID()

		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "could not create request id")
			return
		}
		w.Header().Set("X-Request-ID", requestID)

		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateRequestID() (string, error) {
	b := make([]byte, 16)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
