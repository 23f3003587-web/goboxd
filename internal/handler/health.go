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

const (
	nsjailBinaryPath = "/usr/local/bin/nsjail"
	nsjailConfigPath = "/app/config/nsjail.cfg"
)

var (
	buildVersion  = "0.1.0"
	buildCommit   = "dev-stage1"
	buildDate     = "unknown"
	nsjailVersion = "3.4"
)

func Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		http.Error(w, "encode error", http.StatusInternalServerError)
	}
}

func Readyz(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	languageStatuses, probeErr := getLanguageReadiness()
	readinessErr := checkReadiness()
	if readinessErr != nil || probeErr != nil {
		status = "failed"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	resp := map[string]interface{}{
		"status":    status,
		"nsjail":    getNsjailStatus(),
		"config":    getNsjailConfigStatus(),
		"languages": languageStatuses,
	}

	if readinessErr != nil {
		restErr := map[string]string{"error": readinessErr.Error()}
		resp["error"] = restErr
	}
	if probeErr != nil {
		if existing, ok := resp["error"].(map[string]string); ok {
			existing["languages"] = probeErr.Error()
		} else {
			resp["error"] = map[string]string{"languages": probeErr.Error()}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func Info(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"build_info": map[string]string{
			"version":    buildVersion,
			"commit":     buildCommit,
			"date":       buildDate,
			"go_version": runtime.Version(),
		},
		"nsjail": map[string]interface{}{
			"path":        nsjailBinaryPath,
			"version":     nsjailVersion,
			"config_path": nsjailConfigPath,
			"status":      getNsjailStatus(),
		},
		"languages": buildLanguageInfo(),
		"limits": map[string]int{
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
		log.Printf("Failed to encode response: %v", err)
	}
}

func buildLanguageInfo() []map[string]interface{} {
	ids := make([]string, 0, len(config.Languages))
	for id := range config.Languages {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	languages := make([]map[string]interface{}, 0, len(ids))
	for _, id := range ids {
		lang := config.Languages[id]
		buildLimits := map[string]interface{}{}
		if lang.Build != nil {
			buildLimits = map[string]interface{}{
				"wall_time_s":   lang.Build.Limits.WallTimeS,
				"memory_kb":     lang.Build.Limits.MemoryKB,
				"max_processes": lang.Build.Limits.MaxProcesses,
			}
		}
		languages = append(languages, map[string]interface{}{
			"id":      lang.ID,
			"name":    lang.Name,
			"version": sandbox.LanguageVersion(lang),
			"limits": map[string]interface{}{
				"build": buildLimits,
				"run": map[string]interface{}{
					"wall_time_s":   lang.Run.Limits.WallTimeS,
					"memory_kb":     lang.Run.Limits.MemoryKB,
					"max_processes": lang.Run.Limits.MaxProcesses,
				},
			},
		})
	}
	return languages
}

func getLanguageReadiness() ([]map[string]interface{}, error) {
	ids := make([]string, 0, len(config.Languages))
	for id := range config.Languages {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	statuses := make([]map[string]interface{}, 0, len(ids))
	var failed bool
	for _, id := range ids {
		lang := config.Languages[id]
		status := map[string]interface{}{
			"id":      lang.ID,
			"name":    lang.Name,
			"version": sandbox.LanguageVersion(lang),
			"ready":   false,
		}

		err := sandbox.ProbeLanguage(lang)
		if err != nil {
			status["error"] = err.Error()
			failed = true
		} else {
			status["ready"] = true
		}
		statuses = append(statuses, status)
	}
	if failed {
		return statuses, fmt.Errorf("one or more languages failed smoke probe")
	}
	return statuses, nil
}

func getNsjailStatus() map[string]interface{} {
	status := map[string]interface{}{
		"path":    nsjailBinaryPath,
		"version": nsjailVersion,
		"ok":      false,
	}

	fileInfo, err := os.Stat(nsjailBinaryPath)
	if err != nil {
		status["error"] = err.Error()
		return status
	}

	status["ok"] = fileInfo.Mode().Perm()&0111 != 0
	status["executable"] = status["ok"]
	status["mode"] = fileInfo.Mode().String()
	return status
}

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

func checkReadiness() error {
	nsjailStatus := getNsjailStatus()
	if nsjailStatus["ok"] != true {
		if errValue, ok := nsjailStatus["error"]; ok {
			return fmt.Errorf("nsjail not ready: %v", errValue)
		}
		return fmt.Errorf("nsjail binary is not executable")
	}

	configStatus := getNsjailConfigStatus()
	if configStatus["ok"] != true {
		return fmt.Errorf("nsjail config not ready: %v", configStatus["error"])
	}

	return nil
}
