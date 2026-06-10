package unit_test

import (
	"testing"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/model"
	"github.com/thesouldev/goboxd/internal/security"
)

func TestValidateBuildFlags(t *testing.T) {
	if err := config.LoadLanguages("../../config/languages.yaml"); err != nil {
		t.Fatalf("failed to load languages: %v", err)
	}

	// Valid flags
	if err := security.ValidateBuildFlags("cpp", []string{"-O2", "-Wall", "-std=c++17"}); err != nil {
		t.Fatalf("expected valid build flags to pass, got: %v", err)
	}

	// Disallowed flag
	if err := security.ValidateBuildFlags("cpp", []string{"-fplugin=evil.so"}); err == nil {
		t.Fatal("expected disallowed build flag to return error")
	}
}

func TestValidateRunFlags(t *testing.T) {
	if err := config.LoadLanguages("../../config/languages.yaml"); err != nil {
		t.Fatalf("failed to load languages: %v", err)
	}

	// Most languages have no run flag allowlist
	if err := security.ValidateRunFlags("py3", []string{"-u"}); err == nil {
		t.Fatal("expected run flags to be rejected when no allowlist is configured")
	}
}

func TestValidateRunRequest_Flags(t *testing.T) {
	if err := config.LoadLanguages("../../config/languages.yaml"); err != nil {
		t.Fatalf("failed to load languages: %v", err)
	}

	validReq := model.RunRequest{
		Language: "py3",
		Source:   "print(1)",
		Tests:    []model.TestCase{{Stdin: "", ExpectedStdout: "1"}},
	}

	// Invalid build flag
	invalidBuild := validReq
	invalidBuild.Language = "cpp"
	invalidBuild.Build = &model.BuildRun{Flags: []string{"-fplugin=evil"}}
	if err := security.ValidateRunRequest(invalidBuild); err == nil {
		t.Fatal("expected request with disallowed build flag to be rejected")
	}

	// Invalid filename
	invalidFile := validReq
	invalidFile.SourceFilename = "../../evil.py"
	if err := security.ValidateRunRequest(invalidFile); err == nil {
		t.Fatal("expected request with invalid filename to be rejected")
	}
}
