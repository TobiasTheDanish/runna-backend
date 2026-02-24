package middleware

import (
	"log"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	if !rw.written {
		rw.statusCode = statusCode
		rw.written = true
		rw.ResponseWriter.WriteHeader(statusCode)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

// Logging middleware logs incoming requests and responses
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log incoming request
		log.Printf("[INFO] --> %s %s %s", r.Method, r.URL.Path, r.RemoteAddr)

		// Wrap the response writer to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			written:        false,
		}

		// Call the next handler
		next.ServeHTTP(wrapped, r)

		// Calculate duration
		duration := time.Since(start)

		// Log response with status code
		statusLevel := "INFO"
		if wrapped.statusCode >= 400 && wrapped.statusCode < 500 {
			statusLevel = "WARN"
		} else if wrapped.statusCode >= 500 {
			statusLevel = "ERROR"
		}

		log.Printf("[%s] <-- %s %s %d %s",
			statusLevel,
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration.Round(time.Millisecond))
	})
}
