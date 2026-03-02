package logging

import (
	"log/slog"
	"net/http"
	"time"
)

type responseData struct {
	statusCode int
	method     string
	size       int
	path       string
	duration   time.Duration
}

type responseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if w.responseData.statusCode == 0 {
		w.responseData.statusCode = http.StatusOK
	}
	w.responseData.size += len(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	if w.responseData.statusCode == 0 {
		w.responseData.statusCode = statusCode
		w.ResponseWriter.WriteHeader(statusCode)
	}
}

func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			responseData := &responseData{
				statusCode: 0,
				method:     r.Method,
				size:       0,
				path:       r.URL.Path,
			}
			rw := &responseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}
			next.ServeHTTP(rw, r)
			duration := time.Since(start)
			responseData.duration = duration

			logger.Info("request completed",
				"method", responseData.method,
				"path", responseData.path,
				"status", responseData.statusCode,
				"size", responseData.size,
				"duration", responseData.duration)
		})
	}
}
