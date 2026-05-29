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

// Execute handles the end-to-end sandbox pipeline: writing the file, 
// optional compilation, running all test cases, and aggregating statuses.
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

	srcFilename := lang.SourceFilename
	if req.SourceFilename != "" {
		srcFilename = req.SourceFilename
	}

	if err := jail.WriteSource(srcFilename, req.Source); err != nil {
		return nil, fmt.Errorf("failed to write source: %w", err)
	}

	resp := &model.RunResponse{
		Build: model.BuildResult{Status: "ok"},
	}

	// BUILD STEP (Conditional on language config)
	if lang.Build != nil {
		buildRes, err := runStep(jail, lang.Build, req.Build, srcFilename, lang.Artifact, "")
		resp.Build = *buildRes
		if err != nil || buildRes.Status != "ok" {
			resp.Status = "build_failed"
			resp.Tests = createNotExecutedTests(req.Tests)
			return resp, nil
		}
	}

	// RUN STEP - Sequentially runs every individual test case
	testResults := make([]model.TestResult, 0, len(req.Tests))
	for _, test := range req.Tests {
		runRes, err := runStep(jail, &lang.Run, req.Run, srcFilename, lang.Artifact, test.Stdin)
		if err != nil {
			stderr := ""
			if runRes != nil {
				stderr = runRes.Stderr
			}
			testResults = append(testResults, model.TestResult{
				Status: "runtime_error",
				Stderr: stderr,
			})
			continue
		}

		result := model.TestResult{
			Status:     "accepted",
			Stdout:     runRes.Stdout,
			Stderr:     runRes.Stderr,
			DurationMS: runRes.DurationMS,
		}

		// Comprehensive Status Diff Checker Logic
		rawExpected := strings.TrimSpace(test.ExpectedStdout)
		rawActual := strings.TrimSpace(runRes.Stdout)

		normalize := func(s string) string {
			return strings.Join(strings.Fields(s), " ")
		}

		if rawExpected != "" {
			if rawActual == rawExpected {
				result.Status = "accepted"
			} else if normalize(rawActual) == normalize(rawExpected) {
				result.Status = "output_whitespace_mismatch"
			} else {
				result.Status = "wrong_output"
			}
		}

		testResults = append(testResults, result)
	}

	resp.Tests = testResults
	resp.Status = determineTopLevelStatus(resp.Build, resp.Tests)

	return resp, nil
}

// runStep maps configurations and wraps an execution command within an nsjail process isolation layer
func runStep(jail *Jail, step *config.Step, override *model.BuildRun, srcFilename, artifact, stdin string) (*model.BuildResult, error) {
	// Process template variables in the command
	processedCmd := step.Cmd
	processedCmd = strings.ReplaceAll(processedCmd, "{{source}}", srcFilename)
	processedCmd = strings.ReplaceAll(processedCmd, "{{artifact}}", artifact)

	// Base nsjail configuration
	args := []string{
		"--config", "/app/config/nsjail.cfg",
		"-B", fmt.Sprintf("%s:/app", jail.Path),
		"--time_limit", fmt.Sprintf("%d", step.Limits.WallTimeS),
		"--rlimit_as", fmt.Sprintf("%d", step.Limits.MemoryKB*1024),
		"--quiet",
		"--",
		processedCmd,
	}

	flags := getFlags(override)
	for _, arg := range step.Args {
		arg = strings.ReplaceAll(arg, "{{source}}", srcFilename)
		arg = strings.ReplaceAll(arg, "{{artifact}}", artifact)

		if strings.Contains(arg, "{{flags}}") {
			for _, f := range flags {
				args = append(args, f)
			}
		} else {
			args = append(args, arg)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(step.Limits.WallTimeS)*time.Second+3*time.Second)
	defer cancel()

	start := time.Now()
	cmd := exec.CommandContext(ctx, "/usr/local/bin/nsjail", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	err := cmd.Run()
	duration := int(time.Since(start).Milliseconds())

	status := "ok"
	if err != nil {
		status = "failed"
	}

	return &model.BuildResult{
		Status:     status,
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		DurationMS: duration,
	}, err
}

func getFlags(override *model.BuildRun) []string {
	if override != nil && len(override.Flags) > 0 {
		return override.Flags
	}
	return nil
}

func createNotExecutedTests(tests []model.TestCase) []model.TestResult {
	results := make([]model.TestResult, len(tests))
	for i := range tests {
		results[i] = model.TestResult{Status: "not_executed"}
	}
	return results
}

func determineTopLevelStatus(build model.BuildResult, tests []model.TestResult) string {
	if build.Status != "ok" {
		return "build_failed"
	}
	for _, t := range tests {
		if t.Status != "accepted" {
			return t.Status // Bubble up the first failing test status (e.g., wrong_output, runtime_error)
		}
	}
	return "accepted"
}