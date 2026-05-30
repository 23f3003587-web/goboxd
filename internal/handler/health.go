package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/metrics"
)

// RecoveryMiddleware is a chi-compatible middleware
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("PANIC RECOVERED: %v", rec)
				http.Error(w, `{"error":{"code":"internal_error","message":"Server error"}}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func Readyz(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"status":    "ok",
		"nsjail":    map[string]bool{"ok": true},
		"languages": len(config.Languages),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func Info(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"build_info": map[string]string{
			"version":    "0.1.0",
			"commit":     "dev-stage1",
			"go_version": runtime.Version(),
		},
		"nsjail": map[string]string{
			"path":    "/usr/local/bin/nsjail",
			"version": "3.4",
		},
		"languages": config.Languages,
		"limits": map[string]int{
			"max_source_bytes":    config.Global.MaxSourceBytes,
			"max_tests":           config.Global.MaxTests,
			"max_concurrent_jobs": config.Global.MaxConcurrentJobs,
		},
		"stats": map[string]interface{}{
			"in_flight_jobs":         0,
			"jobs_total":             metrics.Requests(),
			"jobs_failed_internal":   0,
			"last_internal_error_at": time.Now().Format(time.RFC3339),
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
