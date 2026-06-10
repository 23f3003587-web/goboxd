package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// Recovery is the legacy compatibility wrapper used by tests and callers.
func Recovery(next http.Handler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("PANIC RECOVERED: %v", rec)
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("Server error"))
			}
		}()
		next.ServeHTTP(w, r)
	}
}

// Logging writes a concise request summary to the standard logger for tests and diagnostics.
func Logging(next http.Handler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		log.Printf("%s %s from %s - %dB in %s", r.Method, r.URL.Path, r.RemoteAddr, ww.BytesWritten(), time.Since(start).Round(time.Microsecond))
	}
}
