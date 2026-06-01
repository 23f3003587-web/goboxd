package sandbox

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/model"
)

type ExecutionResult struct {
	Status       string
	Stdout       string
	Stderr       string
	DurationMS   int
	MemoryPeakKB int
}

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

	artifact := lang.Artifact
	if req.ArtifactFilename != "" {
		artifact = req.ArtifactFilename
	}

	if err := jail.WriteSource(srcFilename, req.Source); err != nil {
		return nil, fmt.Errorf("failed to write source: %w", err)
	}

	resp := &model.RunResponse{
		Build: model.BuildResult{Status: "ok"},
	}

	if lang.Build != nil {
		buildRes, err := runStep(jail, lang.Build, req.Build, srcFilename, artifact, "")
		resp.Build = model.BuildResult{
			Status:     buildRes.Status,
			Stdout:     buildRes.Stdout,
			Stderr:     buildRes.Stderr,
			DurationMS: buildRes.DurationMS,
		}
		if err != nil {
			return nil, err
		}
		if resp.Build.Status != "ok" {
			resp.Build.Status = "failed"
			resp.Status = "build_failed"
			resp.Tests = createNotExecutedTests(req.Tests)
			return resp, nil
		}
	}

	testResults := make([]model.TestResult, 0, len(req.Tests))
	for _, test := range req.Tests {
		runRes, err := runStep(jail, &lang.Run, req.Run, srcFilename, artifact, test.Stdin)
		if err != nil {
			return nil, err
		}

		result := model.TestResult{
			Stdout:       runRes.Stdout,
			Stderr:       runRes.Stderr,
			DurationMS:   runRes.DurationMS,
			MemoryPeakKB: runRes.MemoryPeakKB,
		}

		if runRes.Status != "ok" {
			switch runRes.Status {
			case "time_exceeded":
				result.Status = "time_exceeded"
			case "memory_exceeded":
				result.Status = "memory_exceeded"
			default:
				result.Status = "runtime_error"
			}
			testResults = append(testResults, result)
			continue
		}

		rawExpected := strings.TrimSpace(test.ExpectedStdout)
		rawActual := strings.TrimSpace(runRes.Stdout)
		normalize := func(s string) string {
			return strings.Join(strings.Fields(s), " ")
		}

		if rawExpected == rawActual {
			result.Status = "accepted"
		} else if normalize(rawActual) == normalize(rawExpected) {
			result.Status = "output_whitespace_mismatch"
		} else {
			result.Status = "wrong_output"
		}

		testResults = append(testResults, result)
	}

	resp.Tests = testResults
	resp.Status = determineTopLevelStatus(resp.Build, resp.Tests)
	return resp, nil
}

// runStep maps configurations and wraps an execution command within an nsjail process isolation layer.
func runStep(jail *Jail, step *config.Step, override *model.BuildRun, srcFilename, artifact, stdin string) (*ExecutionResult, error) {
	processedCmd := processTemplate(step.Cmd, srcFilename, artifact)

	flags := getFlags(override)
	procArgs := make([]string, 0, len(step.Args))
	for _, arg := range step.Args {
		arg = processTemplate(arg, srcFilename, artifact)
		if strings.Contains(arg, "{{flags}}") {
			procArgs = append(procArgs, flags...)
		} else {
			procArgs = append(procArgs, arg)
		}
	}

	limits := mergeLimits(step.Limits, override)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(limits.WallTimeS)*time.Second+3*time.Second)
	defer cancel()

	start := time.Now()

	nsjailPath, lookErr := exec.LookPath("/usr/local/bin/nsjail")
	if lookErr != nil {
		return nil, fmt.Errorf("nsjail binary not found: %w", lookErr)
	}

	nsArgs := []string{
		"--config", "/app/config/nsjail.cfg",
		"-B", fmt.Sprintf("%s:/app", jail.Path),
		"--cwd", "/app",
		"--time_limit", fmt.Sprintf("%d", limits.WallTimeS),
		"--rlimit_as", fmt.Sprintf("%d", limits.MemoryKB*1024),
		"--rlimit_fsize", fmt.Sprintf("%d", config.Global.MaxOutputBytes),
	}
	if limits.MaxProcesses > 0 {
		nsArgs = append(nsArgs, "--rlimit_nproc", fmt.Sprintf("%d", limits.MaxProcesses))
	}
	nsArgs = append(nsArgs,
		"--",
		processedCmd,
	)
	nsArgs = append(nsArgs, procArgs...)
	cmd := exec.CommandContext(ctx, nsjailPath, nsArgs...)

	stdoutWriter := newTruncatingWriter(config.Global.MaxOutputBytes)
	stderrWriter := newTruncatingWriter(config.Global.MaxOutputBytes)
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	err := cmd.Run()
	duration := int(time.Since(start).Milliseconds())
	status := mapExecutionStatus(err, ctx)

	stdoutStr := stdoutWriter.String()
	if stdoutWriter.Truncated() {
		stdoutStr += "\n\n[OUTPUT TRUNCATED - exceeded " + fmt.Sprintf("%d bytes limit", config.Global.MaxOutputBytes) + "]"
	}
	stderrStr := sanitizeNSJailStderr(stderrWriter.String())
	if stderrWriter.Truncated() {
		stderrStr += "\n\n[OUTPUT TRUNCATED - exceeded " + fmt.Sprintf("%d bytes limit", config.Global.MaxOutputBytes) + "]"
	}
	memoryPeak := getMemoryPeakKB(cmd.ProcessState)

	result := &ExecutionResult{
		Status:       status,
		Stdout:       stdoutStr,
		Stderr:       stderrStr,
		DurationMS:   duration,
		MemoryPeakKB: memoryPeak,
	}

	if err != nil {
		if status == "internal_error" {
			return result, err
		}
		return result, nil
	}

	return result, nil
}

func processTemplate(value, source, artifact string) string {
	value = strings.ReplaceAll(value, "{{source}}", source)
	value = strings.ReplaceAll(value, "{{artifact}}", artifact)
	return value
}

func mergeLimits(base config.Limits, override *model.BuildRun) config.Limits {
	if override == nil {
		return base
	}
	if override.Limits.WallTimeS > 0 {
		base.WallTimeS = override.Limits.WallTimeS
	}
	if override.Limits.MemoryKB > 0 {
		base.MemoryKB = override.Limits.MemoryKB
	}
	if override.Limits.MaxProcesses > 0 {
		base.MaxProcesses = override.Limits.MaxProcesses
	}
	return base
}

func mapExecutionStatus(err error, ctx context.Context) string {
	if err == nil {
		return "ok"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "time_exceeded"
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			if status.Signaled() {
				if status.Signal() == syscall.SIGKILL {
					return "memory_exceeded"
				}
				return "runtime_error"
			}
			if status.ExitStatus() != 0 {
				return "runtime_error"
			}
		}
		return "runtime_error"
	}
	return "internal_error"
}

func getMemoryPeakKB(ps *os.ProcessState) int {
	if ps == nil {
		return 0
	}
	if usage, ok := ps.SysUsage().(*syscall.Rusage); ok {
		return int(usage.Maxrss)
	}
	return 0
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
			return t.Status
		}
	}
	return "accepted"
}

// sanitizeNSJailStderr removes known nsjail warning lines that disclose
// internal UID/GID behavior and are noisy for API consumers.
func sanitizeNSJailStderr(s string) string {
	re := regexp.MustCompile(`(?m)^.*(?:\[W\].*logParams\(|Process will be UID/EUID|Process will be GID/EGID).*$\r?\n?`)
	return strings.TrimSpace(re.ReplaceAllString(s, ""))
}

// truncateOutput caps output at maxBytes and adds a truncation marker if exceeded
// (Hole 6 mitigation: prevents OOM from runaway programs)
func truncateOutput(output string, maxBytes int) string {
	if len(output) > maxBytes {
		truncationMarker := "\n\n[OUTPUT TRUNCATED - exceeded " + fmt.Sprintf("%d bytes limit", maxBytes) + "]"
		return output[:maxBytes] + truncationMarker
	}
	return output
}

type truncatingWriter struct {
	buf       bytes.Buffer
	maxBytes  int
	truncated bool
}

func newTruncatingWriter(maxBytes int) *truncatingWriter {
	return &truncatingWriter{maxBytes: maxBytes}
}

func (w *truncatingWriter) Write(p []byte) (int, error) {
	if w.truncated {
		return len(p), nil
	}
	remaining := w.maxBytes - w.buf.Len()
	if remaining <= 0 {
		w.truncated = true
		return len(p), nil
	}
	if len(p) > remaining {
		_, _ = w.buf.Write(p[:remaining])
		w.truncated = true
	} else {
		_, _ = w.buf.Write(p)
	}
	return len(p), nil
}

func (w *truncatingWriter) String() string {
	return w.buf.String()
}

func (w *truncatingWriter) Truncated() bool {
	return w.truncated
}

func LanguageVersion(lang config.Language) string {
	cmdPath := lang.Run.Cmd
	if lang.Build != nil {
		cmdPath = lang.Build.Cmd
	}
	program, err := exec.LookPath(cmdPath)
	if err != nil {
		return "unknown"
	}
	versionCmd := exec.Command(program, "--version")
	output, err := versionCmd.CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(output))
	}
	return strings.TrimSpace(strings.SplitN(string(output), "\n", 2)[0])
}

func ProbeLanguage(lang config.Language) error {
	jail, err := NewJail()
	if err != nil {
		return err
	}
	defer jail.Cleanup()

	source := getProbeSource(lang.SourceFilename)
	if err := jail.WriteSource(lang.SourceFilename, source); err != nil {
		return err
	}

	if lang.Build != nil {
		buildRes, err := runStep(jail, lang.Build, nil, lang.SourceFilename, lang.Artifact, "")
		if err != nil {
			return err
		}
		if buildRes.Status != "ok" {
			return fmt.Errorf("build probe failed: %s", buildRes.Stderr)
		}
	}

	runRes, err := runStep(jail, &lang.Run, nil, lang.SourceFilename, lang.Artifact, "")
	if err != nil {
		return err
	}
	if runRes.Status != "ok" {
		return fmt.Errorf("run probe failed: %s", runRes.Stderr)
	}
	return nil
}

func getProbeSource(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".py":
		return "print('ok')\n"
	case ".cpp", ".cc", ".c":
		return "#include <iostream>\nint main(){std::cout<<\"ok\";return 0;}\n"
	default:
		return "print('ok')\n"
	}
}
