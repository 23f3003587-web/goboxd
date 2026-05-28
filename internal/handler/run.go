package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/thesouldev/goboxd/internal/model"
	"github.com/thesouldev/goboxd/internal/sandbox"
)

func Run(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":{"code":"method_not_allowed","message":"POST only"}}`, http.StatusMethodNotAllowed)
		return
	}

	var req model.RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":{"code":"invalid_json","message":"invalid JSON"}}`, http.StatusBadRequest)
		return
	}

	if req.Language == "" || req.Source == "" || len(req.Tests) == 0 {
		http.Error(w, `{"error":{"code":"invalid_request","message":"language, source and tests are required"}}`, http.StatusBadRequest)
		return
	}

	resp, err := sandbox.Execute(req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":{"code":"execution_error","message":"%s"}}`, err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
