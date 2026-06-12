package unit_test

import (
	"testing"

	"github.com/thesouldev/goboxd/internal/model"
	"github.com/thesouldev/goboxd/internal/sandbox"
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
			name:     "first failing test bubbled up",
			build:    model.BuildResult{Status: "ok"},
			tests:    []model.TestResult{{Status: "accepted"}, {Status: "wrong_output"}, {Status: "runtime_error"}},
			expected: "wrong_output",
		},
		{
			name:     "no tests returns accepted",
			build:    model.BuildResult{Status: "ok"},
			tests:    []model.TestResult{},
			expected: "accepted",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := sandbox.DetermineTopLevelStatus(tt.build, tt.tests)
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
		{"whitespace mismatch", "hello   world", "hello world", "output_whitespace_mismatch"},
		{"wrong output", "hello mars", "hello world", "wrong_output"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			status := sandbox.CompareOutput(tt.actual, tt.expected)
			if status != tt.status {
				t.Fatalf("expected status %s, got %s", tt.status, status)
			}
		})
	}
}

func TestProcessTemplateReplacesPlaceholders(t *testing.T) {
	result := sandbox.ProcessTemplate("{{source}} -o {{artifact}}", "solution.cpp", "solution")
	if result != "solution.cpp -o solution" {
		t.Fatalf("unexpected template result: %s", result)
	}
}

func TestJailWriteSourceRejectsTraversal(t *testing.T) {
	jail, err := sandbox.NewJail()
	if err != nil {
		t.Fatalf("failed to create jail: %v", err)
	}
	defer jail.Cleanup()

	cases := []string{
		"../evil.py",
		"../../etc/passwd",
		"/absolute/path.py",
		"subdir/../../../escape.py",
	}

	for _, name := range cases {
		err = jail.WriteSource(name, "print(1)")
		if err == nil {
			t.Errorf("expected rejection for path %q, but got none", name)
		}
	}
}
