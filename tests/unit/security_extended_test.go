package unit_test

import (
	"strings"
	"testing"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/model"
	"github.com/thesouldev/goboxd/internal/security"
)

func TestValidateRunRequest_RejectsUnknownLanguage(t *testing.T) {
	if err := config.LoadLanguages("../../config/languages.yaml"); err != nil {
		t.Fatalf("failed to load languages: %v", err)
	}

	req := model.RunRequest{
		Language: "not-a-language",
		Source:   "print(1)",
		Tests:    []model.TestCase{{ExpectedStdout: "1"}},
	}

	err := security.ValidateRunRequest(req)
	if err == nil {
		t.Fatal("ValidateRunRequest() error = nil, want unknown-language error")
	}
	if !strings.Contains(err.Error(), "unknown language") {
		t.Fatalf("ValidateRunRequest() error = %v, want unknown-language message", err)
	}
}

func TestValidateRunRequest_RejectsInvalidFilenames(t *testing.T) {
	if err := config.LoadLanguages("../../config/languages.yaml"); err != nil {
		t.Fatalf("failed to load languages: %v", err)
	}

	req := model.RunRequest{
		Language:         "py3",
		Source:           "print(1)",
		SourceFilename:   "../../evil.py",
		ArtifactFilename: "../artifact.out",
		Tests:            []model.TestCase{{ExpectedStdout: "1"}},
	}

	err := security.ValidateRunRequest(req)
	if err == nil {
		t.Fatal("ValidateRunRequest() error = nil, want invalid filename error")
	}
	if !strings.Contains(err.Error(), "invalid filename") {
		t.Fatalf("ValidateRunRequest() error = %v, want invalid filename message", err)
	}
}

func TestValidateBuildFlags_RejectsWhenBuildPhaseIsUnsupported(t *testing.T) {
	if err := config.LoadLanguages("../../config/languages.yaml"); err != nil {
		t.Fatalf("failed to load languages: %v", err)
	}

	err := security.ValidateBuildFlags("py3", []string{"-O2"})
	if err == nil {
		t.Fatal("ValidateBuildFlags() error = nil, want unsupported-build-phase error")
	}
	if !strings.Contains(err.Error(), "does not support build phase") {
		t.Fatalf("ValidateBuildFlags() error = %v, want unsupported-build-phase message", err)
	}
}

func TestValidateRunFlags_RejectsWhenNoAllowlistExists(t *testing.T) {
	if err := config.LoadLanguages("../../config/languages.yaml"); err != nil {
		t.Fatalf("failed to load languages: %v", err)
	}

	err := security.ValidateRunFlags("py3", []string{"-u"})
	if err == nil {
		t.Fatal("ValidateRunFlags() error = nil, want no-allowlist error")
	}
	if !strings.Contains(err.Error(), "does not allow custom run flags") {
		t.Fatalf("ValidateRunFlags() error = %v, want no-allowlist message", err)
	}
}
