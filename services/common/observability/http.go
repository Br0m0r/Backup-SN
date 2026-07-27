package observability

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const RequestIDHeader = "X-Request-ID"

type requestIDKey struct{}

var defaultLogger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

// HTTPLogging adds a request ID and emits one structured log entry per request.
func HTTPLogging(service string, next http.Handler) http.Handler {
	return HTTPLoggingWithLogger(service, defaultLogger, next)
}

// HTTPLoggingWithLogger is HTTPLogging with an injectable logger for tests.
func HTTPLoggingWithLogger(service string, logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		requestID := validOrNewRequestID(r.Header.Get(RequestIDHeader))

		w.Header().Set(RequestIDHeader, requestID)
		r.Header.Set(RequestIDHeader, requestID)
		ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)
		recorder := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(recorder, r.WithContext(ctx))

		logger.Info("http_request",
			"service", service,
			"request_id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.statusCode,
			"bytes", recorder.bytesWritten,
			"duration_ms", float64(time.Since(started).Microseconds())/1000,
		)
	})
}

// RequestIDFromContext returns the request ID installed by HTTPLogging.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(requestIDKey{}).(string)
	return requestID, ok
}

func validOrNewRequestID(candidate string) string {
	candidate = strings.TrimSpace(candidate)
	if candidate != "" && len(candidate) <= 128 {
		valid := true
		for _, char := range candidate {
			if !((char >= 'a' && char <= 'z') ||
				(char >= 'A' && char <= 'Z') ||
				(char >= '0' && char <= '9') ||
				char == '-' || char == '_' || char == '.') {
				valid = false
				break
			}
		}
		if valid {
			return candidate
		}
	}

	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err == nil {
		return hex.EncodeToString(buffer)
	}

	return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (w *responseRecorder) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseRecorder) Write(body []byte) (int, error) {
	written, err := w.ResponseWriter.Write(body)
	w.bytesWritten += written
	return written, err
}

// Unwrap lets http.ResponseController access optional interfaces on the
// underlying writer.
func (w *responseRecorder) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *responseRecorder) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("response writer does not support hijacking")
	}
	return hijacker.Hijack()
}

func (w *responseRecorder) Push(target string, options *http.PushOptions) error {
	pusher, ok := w.ResponseWriter.(http.Pusher)
	if !ok {
		return http.ErrNotSupported
	}
	return pusher.Push(target, options)
}

func (w *responseRecorder) ReadFrom(reader io.Reader) (int64, error) {
	if readerFrom, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		written, err := readerFrom.ReadFrom(reader)
		w.bytesWritten += int(written)
		return written, err
	}
	written, err := io.Copy(struct{ io.Writer }{w}, reader)
	return written, err
}
