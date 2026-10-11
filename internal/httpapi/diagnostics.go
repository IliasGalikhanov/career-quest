package httpapi

import (
	"careerquest/internal/store"
	"log/slog"
	"net/http"
	"time"
)

type observedResponse struct {
	http.ResponseWriter
	status, bytes int
}

func (w *observedResponse) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
func (w *observedResponse) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(200)
	}
	n, err := w.ResponseWriter.Write(body)
	w.bytes += n
	return n, err
}
func (w *observedResponse) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func diagnostics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		id := store.NewID()
		w.Header().Set("X-Request-ID", id)
		observed := &observedResponse{ResponseWriter: w}
		defer func() {
			route := r.Pattern
			if route == "" {
				route = "unmatched"
			}
			status := observed.status
			if status == 0 {
				status = 200
			}
			// Never log query strings, imported content, profile names or reflection text.
			slog.Info("http_request", "request_id", id, "method", r.Method, "route", route, "status", status, "duration_ms", float64(time.Since(start).Microseconds())/1000, "bytes", observed.bytes)
		}()
		next.ServeHTTP(observed, r)
	})
}
