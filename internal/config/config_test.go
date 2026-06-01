package config

import (
	"testing"

	"github.com/thesouldev/goboxd/internal/model"
)

func TestLoadLanguages(t *testing.T) {
	if err := LoadLanguages("../languages.yaml"); err != nil {
		t.Fatalf("failed to load languages: %v", err)
	}
	if _, ok := Languages["py3"]; !ok {
		t.Fatal("expected py3 language to be loaded")
	}
	if _, ok := Languages["cpp"]; !ok {
		t.Fatal("expected cpp language to be loaded")
	}
}

func TestValidateRequest(t *testing.T) {
	if err := LoadLanguages("../languages.yaml"); err != nil {
		t.Fatalf("failed to load languages: %v", err)
	}

	tests := []struct {
		name    string
		req     model.RunRequest
		wantErr bool
	}{
		{
			name: "missing language",
			req: model.RunRequest{
				Source: "print(1)",
				Tests:  []model.TestCase{{Stdin: "", ExpectedStdout: "1"}},
			},
			wantErr: true,
		},
		{
			name: "unknown language",
			req: model.RunRequest{
				Language: "unknown",
				Source:   "print(1)",
				Tests:    []model.TestCase{{Stdin: "", ExpectedStdout: "1"}},
			},
			wantErr: true,
		},
		{
			name: "missing source",
			req: model.RunRequest{
				Language: "py3",
				Tests:    []model.TestCase{{Stdin: "", ExpectedStdout: "1"}},
			},
			wantErr: true,
		},
		{
			name: "valid request",
			req: model.RunRequest{
				Language: "py3",
				Source:   "print(1)",
				Tests:    []model.TestCase{{Stdin: "", ExpectedStdout: "1"}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
