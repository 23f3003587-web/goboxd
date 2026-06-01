package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/metrics"
	"github.com/thesouldev/goboxd/internal/model"
	"github.com/thesouldev/goboxd/internal/sandbox"
	"github.com/thesouldev/goboxd/internal/security"
)

func Run(w http.ResponseWriter, r *http.Request) {
	// 1. Enforce strict POST method routing
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte(`{"error":{"code":"method_not_allowed","message":"POST only"}}`))
		return
	}

	// 2. Decode incoming JSON safely and enforce a total request size cap at the HTTP layer.
	r.Body = http.MaxBytesReader(w, r.Body, int64(config.Global.MaxRequestBytes))
	var req model.RunRequest
	dec := json.NewDecoder(r.Body)
	// reject unknown fields to avoid silently ignoring attacker-supplied keys
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"code":"invalid_json","message":"invalid JSON"}}`))
		return
	}

	// 3. Run validation without consuming concurrency.
	if err := security.ValidateRunRequest(req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		respErr := map[string]map[string]string{
			"error": {
				"code":    "validation_error",
				"message": err.Error(),
			},
		}
		_ = json.NewEncoder(w).Encode(respErr)
		return
	}

	// 4. Enforce the global concurrency limit and queue when full.
	if err := AcquireJob(r.Context()); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		respErr := map[string]map[string]string{
			"error": {
				"code":    "server_busy",
				"message": "server is processing maximum concurrent jobs",
			},
		}
		_ = json.NewEncoder(w).Encode(respErr)
		return
	}
	defer ReleaseJob()

	metrics.IncRequests()
	// Defensive re-check before executing the job to ensure validation
	// wasn't bypassed earlier (extra safety for running instances).
	if err := security.ValidateRunRequest(req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		respErr := map[string]map[string]string{
			"error": {
				"code":    "validation_error",
				"message": err.Error(),
			},
		}
		_ = json.NewEncoder(w).Encode(respErr)
		return
	}

	resp, err := sandbox.Execute(req)
	if err != nil {
		metrics.IncFailedInternal()
		metrics.SetLastInternalErrorAt(time.Now().Format(time.RFC3339))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		respErr := map[string]map[string]string{
			"error": {
				"code":    "internal_error",
				"message": err.Error(),
			},
		}
		_ = json.NewEncoder(w).Encode(respErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
