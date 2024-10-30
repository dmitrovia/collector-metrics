package requestmiddleware

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

type ResponseWriter interface {
	Header() http.Header
	Write(b []byte) (int, error)
	WriteHeader(statusCode int)
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size

	return size, fmt.Errorf("loggingResponseWriterWrite: %w", err)
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func RequestLogger(zapLogger *zap.Logger) func(http.Handler) http.Handler {
	handler := func(hand http.Handler) http.Handler {
		logFn := func(writer http.ResponseWriter, req *http.Request) {
			start := time.Now()

			responseData := &responseData{
				status: 0,
				size:   0,
			}

			lw := loggingResponseWriter{
				ResponseWriter: writer,
				responseData:   responseData,
			}

			hand.ServeHTTP(&lw, req)

			duration := time.Since(start)

			zapLogger.Info("got incoming HTTP request",
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path),
				zap.Duration("duration", duration),
				zap.Int("status", responseData.status),
				zap.Int("size", responseData.size),
			)
		}

		return http.HandlerFunc(logFn)
	}

	return handler
}
