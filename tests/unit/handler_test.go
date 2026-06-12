package unit_test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	handler "github.com/thesouldev/goboxd/internal/api/handlers"
)

func TestRecoveryReturnsInternalServerError(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)

	handler.Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "Server error") {
		t.Fatalf("expected panic recovery response to mention server error, got %q", recorder.Body.String())
	}
}

func TestLoggingWritesRequestSummary(t *testing.T) {
	var buf bytes.Buffer
	previousWriter := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(previousWriter)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	req.RemoteAddr = "127.0.0.1:4321"

	handler.Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}

	logLine := buf.String()
	if !strings.Contains(logLine, "GET /ok") || !strings.Contains(logLine, "127.0.0.1") {
		t.Fatalf("expected logging middleware to write request details, got %q", logLine)
	}
}

func TestRecoveryMiddlewareReturnsStructuredError(t *testing.T) {
	middleware := handler.RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)

	middleware.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}

	var payload map[string]any
	if err := json.NewDecoder(strings.NewReader(recorder.Body.String())).Decode(&payload); err != nil {
		t.Fatalf("expected valid JSON response, got error: %v", err)
	}

	errorPayload, ok := payload["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error payload to be an object, got %T", payload["error"])
	}

	if errorPayload["code"] != "internal_error" || errorPayload["message"] != "internal server error" {
		t.Fatalf("unexpected error payload: %#v", errorPayload)
	}
}

func TestHealthzReturnsOKPayload(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	handler.Healthz(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	if got := recorder.Header().Get("Content-Type"); !strings.Contains(got, "application/json") {
		t.Fatalf("expected JSON content type, got %q", got)
	}

	var payload map[string]string
	if err := json.NewDecoder(strings.NewReader(recorder.Body.String())).Decode(&payload); err != nil {
		t.Fatalf("expected valid JSON response, got error: %v", err)
	}

	if payload["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", payload["status"])
	}
}
