package config

import (
	"fmt"
	"time"

	"github.com/thesouldev/goboxd/internal/model"
)

type GlobalConfig struct {
	MaxSourceBytes    int           `yaml:"max_source_bytes"`
	MaxTests          int           `yaml:"max_tests"`
	MaxConcurrentJobs int           `yaml:"max_concurrent_jobs"`
	RequestTimeout    time.Duration `yaml:"request_timeout"`
}

var Global = GlobalConfig{
	MaxSourceBytes:    256 * 1024, // 256 KiB
	MaxTests:          50,
	MaxConcurrentJobs: 16,
	RequestTimeout:    30 * time.Second,
}

func LoadConfig() error {
	// TODO: Load from YAML later if needed
	return nil
}

func ValidateRequest(req model.RunRequest) error {
	if len(req.Source) > Global.MaxSourceBytes {
		return fmt.Errorf("source too large: max %d bytes", Global.MaxSourceBytes)
	}
	if len(req.Tests) == 0 || len(req.Tests) > Global.MaxTests {
		return fmt.Errorf("invalid number of tests: must be 1-%d", Global.MaxTests)
	}
	if req.Language == "" {
		return fmt.Errorf("language is required")
	}
	return nil
}
