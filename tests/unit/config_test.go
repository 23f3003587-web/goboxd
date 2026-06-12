package unit_test

import (
	"testing"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/model"
)

func TestLoadLanguages(t *testing.T) {
	if err := config.LoadLanguages("../../config/languages.yaml"); err != nil {
		t.Fatalf("failed to load languages: %v", err)
	}
	if _, ok := config.GetLanguage("py3"); !ok {
		t.Fatal("expected py3 language to be loaded")
	}
}

func TestValidateRequest(t *testing.T) {
	if err := config.LoadLanguages("../../config/languages.yaml"); err != nil {
		t.Fatalf("failed to load languages: %v", err)
	}

	tests := []struct {
		name    string
		req     model.RunRequest
		wantErr bool
	}{
		{"missing language", model.RunRequest{Source: "print(1)", Tests: []model.TestCase{{Stdin: "", ExpectedStdout: "1"}}}, true},
		{"unknown language", model.RunRequest{Language: "unknown", Source: "print(1)", Tests: []model.TestCase{{Stdin: "", ExpectedStdout: "1"}}}, true},
		{"missing source", model.RunRequest{Language: "py3", Tests: []model.TestCase{{Stdin: "", ExpectedStdout: "1"}}}, true},
		{"valid request", model.RunRequest{Language: "py3", Source: "print(1)", Tests: []model.TestCase{{Stdin: "", ExpectedStdout: "1"}}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.ValidateRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
