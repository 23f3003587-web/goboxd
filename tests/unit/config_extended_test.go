package unit_test

import (
	"os"
	"strings"
	"testing"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/model"
)

func TestLoadConfig_UsesEnvironmentValue(t *testing.T) {
	oldValue, hadValue := os.LookupEnv("MAX_CONCURRENT_JOBS")
	original := config.Global.MaxConcurrentJobs

	t.Cleanup(func() {
		if hadValue {
			_ = os.Setenv("MAX_CONCURRENT_JOBS", oldValue)
		} else {
			_ = os.Unsetenv("MAX_CONCURRENT_JOBS")
		}
		config.Global.MaxConcurrentJobs = original
	})

	_ = os.Setenv("MAX_CONCURRENT_JOBS", "7")
	config.Global.MaxConcurrentJobs = 1

	if err := config.LoadConfig(); err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if got := config.Global.MaxConcurrentJobs; got != 7 {
		t.Fatalf("LoadConfig() set MaxConcurrentJobs = %d, want 7", got)
	}
}

func TestLoadConfig_RejectsInvalidEnvironmentValue(t *testing.T) {
	oldValue, hadValue := os.LookupEnv("MAX_CONCURRENT_JOBS")
	original := config.Global.MaxConcurrentJobs

	t.Cleanup(func() {
		if hadValue {
			_ = os.Setenv("MAX_CONCURRENT_JOBS", oldValue)
		} else {
			_ = os.Unsetenv("MAX_CONCURRENT_JOBS")
		}
		config.Global.MaxConcurrentJobs = original
	})

	_ = os.Setenv("MAX_CONCURRENT_JOBS", "not-a-number")
	config.Global.MaxConcurrentJobs = 1

	err := config.LoadConfig()
	if err == nil {
		t.Fatal("LoadConfig() error = nil, want invalid MAX_CONCURRENT_JOBS error")
	}
	if !strings.Contains(err.Error(), "invalid MAX_CONCURRENT_JOBS") {
		t.Fatalf("LoadConfig() error = %v, want invalid MAX_CONCURRENT_JOBS message", err)
	}
}

func TestValidateRequest_RejectsOversizedInput(t *testing.T) {
	if err := config.LoadLanguages("../../config/languages.yaml"); err != nil {
		t.Fatalf("failed to load languages: %v", err)
	}

	original := config.Global
	t.Cleanup(func() { config.Global = original })

	config.Global.MaxSourceBytes = 5
	config.Global.MaxTests = 10
	config.Global.MaxTestStdinBytes = 1024
	config.Global.MaxExpectedStdoutBytes = 1024

	req := model.RunRequest{
		Language: "py3",
		Source:   "123456",
		Tests:    []model.TestCase{{ExpectedStdout: "ok"}},
	}

	err := config.ValidateRequest(req)
	if err == nil {
		t.Fatal("ValidateRequest() error = nil, want source too large error")
	}
	if !strings.Contains(err.Error(), "source too large") {
		t.Fatalf("ValidateRequest() error = %v, want source too large message", err)
	}
}
