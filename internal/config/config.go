package config

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/thesouldev/goboxd/internal/model"
)

type GlobalConfig struct {
	MaxSourceBytes         int           `yaml:"max_source_bytes"`
	MaxTests               int           `yaml:"max_tests"`
	MaxTestStdinBytes      int           `yaml:"max_test_stdin_bytes"`
	MaxExpectedStdoutBytes int           `yaml:"max_expected_stdout_bytes"`
	MaxRequestBytes        int           `yaml:"max_request_bytes"`
	MaxConcurrentJobs      int           `yaml:"max_concurrent_jobs"`
	RequestTimeout         time.Duration `yaml:"request_timeout"`
	MaxOutputBytes         int           `yaml:"max_output_bytes"`
}

var (
	Global GlobalConfig
	once   sync.Once
)

func init() {
	Global = GlobalConfig{
		MaxSourceBytes:         256 * 1024, // 256 KiB
		MaxTests:               50,
		MaxTestStdinBytes:      16 * 1024,
		MaxExpectedStdoutBytes: 64 * 1024,
		MaxRequestBytes:        2 * 1024 * 1024, // 2 MiB
		MaxConcurrentJobs: func() int {
			n := runtime.NumCPU()
			if n < 1 {
				return 1
			}
			return n
		}(),
		RequestTimeout: 30 * time.Second,
		MaxOutputBytes: 1024 * 1024, // 1 MiB
	}
}

func LoadConfig() error {
	once.Do(func() {
		// Future: load from YAML/config file here if needed
	})

	if env := os.Getenv("MAX_CONCURRENT_JOBS"); env != "" {
		value, err := strconv.Atoi(env)
		if err != nil {
			return fmt.Errorf("invalid MAX_CONCURRENT_JOBS: %w", err)
		}
		if value <= 0 {
			return fmt.Errorf("MAX_CONCURRENT_JOBS must be greater than zero")
		}
		Global.MaxConcurrentJobs = value
	}

	if Global.MaxConcurrentJobs < 1 {
		Global.MaxConcurrentJobs = 1
	}

	return nil
}

// GetConfig returns a copy to avoid any mutation risks
func GetConfig() GlobalConfig {
	once.Do(func() { _ = LoadConfig() }) // ensure loaded
	return Global
}

func ValidateRequest(req model.RunRequest) error {
	if req.Language == "" {
		return fmt.Errorf("language is required")
	}
	if _, ok := GetLanguage(req.Language); !ok {
		return fmt.Errorf("unknown language: %s", req.Language)
	}
	if req.Source == "" {
		return fmt.Errorf("source is required")
	}
	if len(req.Source) > Global.MaxSourceBytes {
		return fmt.Errorf("source too large: max %d bytes", Global.MaxSourceBytes)
	}
	if len(req.Tests) == 0 || len(req.Tests) > Global.MaxTests {
		return fmt.Errorf("invalid number of tests: must be 1-%d", Global.MaxTests)
	}

	for idx, test := range req.Tests {
		if len(test.Stdin) > Global.MaxTestStdinBytes {
			return fmt.Errorf("test[%d].stdin too large: max %d bytes", idx, Global.MaxTestStdinBytes)
		}
		if len(test.ExpectedStdout) > Global.MaxExpectedStdoutBytes {
			return fmt.Errorf("test[%d].expected_stdout too large: max %d bytes", idx, Global.MaxExpectedStdoutBytes)
		}
	}
	return nil
}
