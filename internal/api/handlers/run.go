package handlers

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
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte(`{"error":{"code":"method_not_allowed","message":"POST only"}}`))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, int64(config.Global.MaxRequestBytes))
	var req model.RunRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"code":"invalid_json","message":"invalid JSON"}}`))
		return
	}

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
