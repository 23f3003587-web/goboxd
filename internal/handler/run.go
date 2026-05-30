package handler

import (
	"encoding/json"
	"net/http"

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

	// 2. Decode incoming JSON safely
	var req model.RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"code":"invalid_json","message":"invalid JSON"}}`))
		return
	}

	// 3. Execute consolidated security validations (Hole mitigations + Global constraints)
	if err := security.ValidateRunRequest(req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		// Return structured validation error text dynamically
		respErr := map[string]map[string]string{
			"error": {
				"code":    "validation_error",
				"message": err.Error(),
			},
		}
		_ = json.NewEncoder(w).Encode(respErr)
		return
	}

	// 4. Pass execution to the containerized sandbox environment
	resp, err := sandbox.Execute(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		respErr := map[string]map[string]string{
			"error": {
				"code":    "execution_error",
				"message": err.Error(),
			},
		}
		_ = json.NewEncoder(w).Encode(respErr)
		return
	}

	// 5. Send successful execution metrics back to the client
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
