package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/model"
)

func Execute(req model.RunRequest) (*model.RunResponse, error) {
	lang, exists := config.GetLanguage(req.Language)
	if !exists {
		return nil, fmt.Errorf("unknown language: %s", req.Language)
	}

	jail, err := NewJail()
	if err != nil {
		return nil, fmt.Errorf("failed to create jail: %w", err)
	}
	defer jail.Cleanup()

	// Determine source filename
	srcFilename := lang.SourceFilename
	if req.SourceFilename != "" {
		srcFilename = req.SourceFilename
	}

	if err := jail.WriteSource(srcFilename, req.Source); err != nil {
		return nil, fmt.Errorf("failed to write source: %w", err)
	}

	// Build nsjail arguments
	args := []string{
		"--mode", "o",
		"--time_limit", fmt.Sprintf("%d", lang.Run.Limits.WallTimeS),
		"--rlimit_as", fmt.Sprintf("%d", lang.Run.Limits.MemoryKB),
		"-R", "/usr",
		"-R", "/lib",
		"-R", "/lib64",
		"-R", "/bin",
		"-B", fmt.Sprintf("%s:/app", jail.Path),
		"--chroot", "/app",
		"--", lang.Run.Cmd,
	}

	// Fix: Use relative path inside chroot (important!)
	for _, arg := range lang.Run.Args {
		arg = strings.ReplaceAll(arg, "{{source}}", srcFilename)
		// Ensure no leading slash for files inside chroot
		if !strings.HasPrefix(arg, "/app/") {
        arg = "/app/" + strings.TrimPrefix(arg, "/")
    }
    
    args = append(args, arg)
    }

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(lang.Run.Limits.WallTimeS)*time.Second+2*time.Second)
	defer cancel()

	start := time.Now()
	cmd := exec.CommandContext(ctx, "/usr/local/bin/nsjail", args...)
	cmd.Dir = jail.Path

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	duration := int(time.Since(start).Milliseconds())

	testStatus := "accepted"
	stdoutStr := stdout.String()

	if len(req.Tests) > 0 {
		expected := strings.TrimSpace(req.Tests[0].ExpectedStdout)
		actual := strings.TrimSpace(stdoutStr)
		if expected != "" && !strings.Contains(actual, expected) {
			testStatus = "wrong_output"
		}
	}

	if err != nil && testStatus == "accepted" {
		testStatus = "runtime_error"
	}

	resp := &model.RunResponse{
		Status: testStatus,
		Build: model.BuildResult{
			Status:     "ok",
			Stdout:     "",
			Stderr:     stderr.String(),
			DurationMS: duration,
		},
		Tests: []model.TestResult{{
			Status:       testStatus,
			Stdout:       stdoutStr,
			Stderr:       stderr.String(),
			DurationMS:   duration,
			MemoryPeakKB: 0,
		}},
	}
	return resp, nil
}
