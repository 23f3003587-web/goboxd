package sandbox

import (
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/thesouldev/goboxd/internal/model"
)

func TestDetermineTopLevelStatus(t *testing.T) {
	cases := []struct {
		name     string
		build    model.BuildResult
		tests    []model.TestResult
		expected string
	}{
		{
			name:     "accepted when all accepted",
			build:    model.BuildResult{Status: "ok"},
			tests:    []model.TestResult{{Status: "accepted"}, {Status: "accepted"}},
			expected: "accepted",
		},
		{
			name:     "build failed returns build_failed",
			build:    model.BuildResult{Status: "failed"},
			tests:    []model.TestResult{{Status: "accepted"}},
			expected: "build_failed",
		},
		{
			name:     "first failing test bubbled",
			build:    model.BuildResult{Status: "ok"},
			tests:    []model.TestResult{{Status: "accepted"}, {Status: "wrong_output"}, {Status: "runtime_error"}},
			expected: "wrong_output",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := determineTopLevelStatus(tt.build, tt.tests)
			if result != tt.expected {
				t.Fatalf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestWhitespaceMismatchComparison(t *testing.T) {
	cases := []struct {
		name     string
		actual   string
		expected string
		status   string
	}{
		{"exact match", "hello world\n", "hello world", "accepted"},
		{"whitespace mismatch", "hello   world  ", "hello world", "output_whitespace_mismatch"},
		{"wrong output", "hello mars", "hello world", "wrong_output"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			rawActual := strings.TrimSpace(tt.actual)
			rawExpected := strings.TrimSpace(tt.expected)
			normalize := func(s string) string {
				return strings.Join(strings.Fields(s), " ")
			}

			var status string
			if rawActual == rawExpected {
				status = "accepted"
			} else if normalize(rawActual) == normalize(rawExpected) {
				status = "output_whitespace_mismatch"
			} else {
				status = "wrong_output"
			}

			if status != tt.status {
				t.Fatalf("expected status %s, got %s", tt.status, status)
			}
		})
	}
}

func TestProcessTemplateReplacesPlaceholders(t *testing.T) {
	result := processTemplate("{{source}} -o {{artifact}}", "solution.cpp", "solution")
	if result != "solution.cpp -o solution" {
		t.Fatalf("unexpected template result: %s", result)
	}
}

func TestMapExecutionStatus(t *testing.T) {
	ctx := context.Background()

	if got := mapExecutionStatus(context.DeadlineExceeded, ctx); got != "time_exceeded" {
		t.Fatalf("expected time_exceeded, got %s", got)
	}

	if err := exec.Command("false").Run(); err != nil {
		if got := mapExecutionStatus(err, ctx); got != "runtime_error" {
			t.Fatalf("expected runtime_error, got %s", got)
		}
	} else {
		t.Fatal("expected false command to fail")
	}

	cmd := exec.Command("sh", "-c", "kill -9 $$")
	if err := cmd.Run(); err != nil {
		if got := mapExecutionStatus(err, ctx); got != "memory_exceeded" {
			t.Fatalf("expected memory_exceeded, got %s", got)
		}
	} else {
		t.Fatal("expected signaled command to fail")
	}
}

func TestJailWriteSourceRejectsTraversal(t *testing.T) {
	jail, err := NewJail()
	if err != nil {
		t.Fatalf("failed to create jail: %v", err)
	}
	defer jail.Cleanup()

	if err := jail.WriteSource("../evil.py", "print(1)"); err == nil {
		t.Fatal("expected path traversal in filename to be rejected")
	}
}
