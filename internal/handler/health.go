package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sort"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/metrics"
	"github.com/thesouldev/goboxd/internal/sandbox"
)

// ── Constants ────────────────────────────────────────────────────────────────

const (
	nsjailBinaryPath = "/usr/local/bin/nsjail"
	nsjailConfigPath = "/app/config/nsjail.cfg"
)

var (
	buildVersion  = "0.2.0-stage2"
	buildCommit   = "dev"
	buildDate     = "unknown"
	nsjailVersion = "3.4"
)

// ── Middleware ────────────────────────────────────────────────────────────────

// RecoveryMiddleware catches panics and returns a clean 500.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("PANIC RECOVERED: %v", rec)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"error": map[string]string{
						"code":    "internal_error",
						"message": "internal server error",
					},
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ── Healthz (/healthz) ────────────────────────────────────────────────────────
// Lightweight liveness probe — just confirms the process is alive.

func Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ── Readyz (/readyz) ──────────────────────────────────────────────────────────
// Deep readiness probe — checks nsjail binary, config, and every language.
// Returns 200 if fully ready, 503 if any check fails.

func Readyz(w http.ResponseWriter, r *http.Request) {
	nsjailStatus := getNsjailStatus()
	configStatus := getNsjailConfigStatus()
	langStatuses, langErr := probeAllLanguages()

	allReady := nsjailStatus["ok"] == true &&
		configStatus["ok"] == true &&
		langErr == nil

	httpStatus := http.StatusOK
	overallStatus := "ok"
	if !allReady {
		httpStatus = http.StatusServiceUnavailable
		overallStatus = "degraded"
	}

	resp := map[string]interface{}{
		"status":    overallStatus,
		"nsjail":    nsjailStatus,
		"config":    configStatus,
		"languages": langStatuses,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("readyz: encode error: %v", err)
	}
}

// ── Info (/info) ──────────────────────────────────────────────────────────────
// Returns build metadata, nsjail info, all language configs, global limits,
// and live runtime stats.

func Info(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"build": map[string]string{
			"version":    buildVersion,
			"commit":     buildCommit,
			"date":       buildDate,
			"go_version": runtime.Version(),
		},
		"nsjail": map[string]interface{}{
			"path":        nsjailBinaryPath,
			"version":     nsjailVersion,
			"config_path": nsjailConfigPath,
			"ok":          getNsjailStatus()["ok"],
		},
		"languages": buildLanguageInfo(),
		"limits": map[string]interface{}{
			"max_source_bytes":    config.Global.MaxSourceBytes,
			"max_tests":           config.Global.MaxTests,
			"max_concurrent_jobs": config.Global.MaxConcurrentJobs,
			"max_output_bytes":    config.Global.MaxOutputBytes,
		},
		"stats": map[string]interface{}{
			"in_flight_jobs":         metrics.InFlight(),
			"jobs_total":             metrics.Requests(),
			"jobs_failed_internal":   metrics.FailedInternal(),
			"last_internal_error_at": metrics.LastInternalErrorAt(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("info: encode error: %v", err)
	}
}

// ── Internal helpers ──────────────────────────────────────────────────────────

// getNsjailStatus checks the nsjail binary exists and is executable.
func getNsjailStatus() map[string]interface{} {
	status := map[string]interface{}{
		"path": nsjailBinaryPath,
		"ok":   false,
	}
	fi, err := os.Stat(nsjailBinaryPath)
	if err != nil {
		status["error"] = err.Error()
		return status
	}
	executable := fi.Mode().Perm()&0111 != 0
	status["ok"] = executable
	status["executable"] = executable
	status["mode"] = fi.Mode().String()
	return status
}

// getNsjailConfigStatus checks the nsjail config file exists.
func getNsjailConfigStatus() map[string]interface{} {
	status := map[string]interface{}{
		"path": nsjailConfigPath,
		"ok":   false,
	}
	if _, err := os.Stat(nsjailConfigPath); err != nil {
		status["error"] = err.Error()
		return status
	}
	status["ok"] = true
	return status
}

// probeAllLanguages runs a smoke test for every registered language.
// Returns per-language status map and a top-level error if any language fails.
func probeAllLanguages() ([]map[string]interface{}, error) {
	ids := sortedLanguageIDs()
	statuses := make([]map[string]interface{}, 0, len(ids))
	anyFailed := false

	for _, id := range ids {
		lang := config.Languages[id]
		entry := map[string]interface{}{
			"id":      lang.ID,
			"name":    lang.Name,
			"version": sandbox.LanguageVersion(lang),
			"ready":   false,
		}
		if err := sandbox.ProbeLanguage(lang); err != nil {
			entry["error"] = err.Error()
			anyFailed = true
		} else {
			entry["ready"] = true
		}
		statuses = append(statuses, entry)
	}

	if anyFailed {
		return statuses, fmt.Errorf("one or more languages failed readiness probe")
	}
	return statuses, nil
}

// buildLanguageInfo builds the language list for /info — includes limits and version.
func buildLanguageInfo() []map[string]interface{} {
	ids := sortedLanguageIDs()
	result := make([]map[string]interface{}, 0, len(ids))

	for _, id := range ids {
		lang := config.Languages[id]

		entry := map[string]interface{}{
			"id":      lang.ID,
			"name":    lang.Name,
			"version": sandbox.LanguageVersion(lang),
			"run_limits": map[string]interface{}{
				"wall_time_s":   lang.Run.Limits.WallTimeS,
				"memory_kb":     lang.Run.Limits.MemoryKB,
				"max_processes": lang.Run.Limits.MaxProcesses,
			},
		}

		if lang.Build != nil {
			entry["build_limits"] = map[string]interface{}{
				"wall_time_s":   lang.Build.Limits.WallTimeS,
				"memory_kb":     lang.Build.Limits.MemoryKB,
				"max_processes": lang.Build.Limits.MaxProcesses,
			}
			if len(lang.Build.FlagAllowlist) > 0 {
				entry["build_flag_allowlist"] = lang.Build.FlagAllowlist
			}
		}

		result = append(result, entry)
	}
	return result
}

// sortedLanguageIDs returns language IDs in alphabetical order for stable output.
func sortedLanguageIDs() []string {
	ids := make([]string, 0, len(config.Languages))
	for id := range config.Languages {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
