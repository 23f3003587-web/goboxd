package integration

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/handler"
)

func TestRunEndpoint(t *testing.T) {
	config.LoadLanguages("languages.yaml")

	tests := []struct {
		name       string
		req        string
		wantStatus int
		wantCode   string
	}{
		{
			name: "python hello",
			req: `{"language":"py3","source":"print('Hello World')","tests":[{"stdin":"","expected_stdout":"Hello World"}]}`,
			wantStatus: 200,
		},
		{
			name: "bad filename",
			req: `{"language":"py3","source_filename":"../../bad","source":"print(1)","tests":[{"stdin":"","expected_stdout":"1"}]}`,
			wantStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/run", bytes.NewBufferString(tt.req))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.Run(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("handler returned wrong status: got %v want %v; body=%s", rr.Code, tt.wantStatus, rr.Body.String())
			}
		})
	}
}
