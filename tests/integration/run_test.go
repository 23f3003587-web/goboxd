// TestRun_AllLanguages exercises the /run endpoint for every registered language in config/languages.yaml.
// It verifies the happy path for each language, plus wrong-output and invalid-language handling.
package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/model"
)

const defaultServerURL = "http://localhost:8080"

func TestRun_AllLanguages(t *testing.T) {
	baseURL := serverURL(t)
	requireServerReachable(t, baseURL)

	languages := loadLanguages(t)

	for langID, lang := range languages {
		langID, lang := langID, lang
		t.Run(langID, func(t *testing.T) {
			req := buildHelloWorldRequest(langID, lang)
			resp := sendRunRequest(t, baseURL, req)

			require.NotEmpty(t, resp.Status, "top-level status should be present")
			require.Equal(t, "accepted", resp.Status, "expected successful run to be accepted")
			require.Equal(t, "ok", resp.Build.Status, "expected build phase to succeed")
			require.NotEmpty(t, resp.Tests, "expected at least one test result")

			testResult := resp.Tests[0]
			require.Equal(t, "accepted", testResult.Status, "expected hello-world test to be accepted")
			assert.NotEmpty(t, testResult.Stdout, "expected stdout to be returned")
			assert.GreaterOrEqual(t, testResult.DurationMS, 0, "duration must be non-negative")
			assert.GreaterOrEqual(t, testResult.MemoryPeakKB, 0, "memory peak must be non-negative")
			assert.GreaterOrEqual(t, resp.Build.DurationMS, 0, "build duration must be non-negative")
		})
	}
}

func TestRun_FailureModes(t *testing.T) {
	baseURL := serverURL(t)
	requireServerReachable(t, baseURL)

	t.Run("wrong_output", func(t *testing.T) {
		req := buildHelloWorldRequest("py3", config.Languages["py3"])
		req.Tests = []model.TestCase{{Stdin: "", ExpectedStdout: "definitely not the expected output"}}

		resp := sendRunRequest(t, baseURL, req)
		require.Equal(t, "wrong_output", resp.Status, "expected wrong-output status for mismatched expected stdout")
		require.NotEmpty(t, resp.Tests)
		require.Equal(t, "wrong_output", resp.Tests[0].Status)
	})

	t.Run("invalid_language", func(t *testing.T) {
		req := model.RunRequest{
			Language: "definitely-not-a-language",
			Source:   "print('hello')",
			Tests:    []model.TestCase{{Stdin: "", ExpectedStdout: "hello"}},
		}

		statusCode, body := postRunExpectStatus(t, baseURL, req, http.StatusBadRequest)
		require.Equal(t, http.StatusBadRequest, statusCode)
		assert.True(t, strings.Contains(string(body), "language") || strings.Contains(string(body), "unknown"), string(body))
	})
}

func loadLanguages(t *testing.T) map[string]config.Language {
	t.Helper()

	repoRoot := findRepoRoot(t)
	require.NoError(t, config.LoadLanguages(filepath.Join(repoRoot, "config", "languages.yaml")), "failed to load config/languages.yaml")

	return config.Languages
}

func buildHelloWorldRequest(langID string, lang config.Language) model.RunRequest {
	req := model.RunRequest{
		Language: langID,
		Tests: []model.TestCase{{
			Stdin:          "",
			ExpectedStdout: expectedHelloOutput(langID),
		}},
	}

	req.Source = helloWorldSource(langID)

	if lang.SourceFilename != "" {
		req.SourceFilename = lang.SourceFilename
	}
	if lang.Artifact != "" {
		req.ArtifactFilename = lang.Artifact
	}

	switch langID {
	case "java":
		req.SourceFilename = "Hello.java"
		req.ArtifactFilename = "Hello"
	case "cpp", "c":
		req.ArtifactFilename = "solution"
	case "verilog":
		req.SourceFilename = "solution.v"
		req.ArtifactFilename = "solution"
	}

	// Build and Run are value types (not pointers)
	if lang.Build != nil && lang.Build.Cmd != "" {
		req.Build = &model.BuildRun{Limits: toModelLimits(lang.Build.Limits)}
	}
	if lang.Run.Cmd != "" {
		req.Run = &model.BuildRun{Limits: toModelLimits(lang.Run.Limits)}
	}

	return req
}

func helloWorldSource(langID string) string {
	switch langID {
	case "c":
		return "#include <stdio.h>\n\nint main(void) { puts(\"Hello from c\"); return 0; }"
	case "cpp":
		return "#include <iostream>\n\nint main() { std::cout << \"Hello from cpp\\n\"; return 0; }"
	case "java":
		return "public class Hello { public static void main(String[] args) { System.out.println(\"Hello from java\"); } }"
	case "js":
		return "console.log('Hello from js');"
	case "bash":
		return "echo 'Hello from bash'"
	case "verilog":
		return "module top; initial begin $display(\"Hello from verilog\"); $finish; end endmodule"
	case "lua":
		return "print('Hello from lua')"
	default:
		return "print('Hello from py3')"
	}
}

func toModelLimits(limits config.Limits) model.Limits {
	return model.Limits{
		WallTimeS:    limits.WallTimeS,
		MemoryKB:     limits.MemoryKB,
		MaxProcesses: limits.MaxProcesses,
	}
}

func expectedHelloOutput(langID string) string {
	switch langID {
	case "c":
		return "Hello from c"
	case "cpp":
		return "Hello from cpp"
	case "java":
		return "Hello from java"
	case "js":
		return "Hello from js"
	case "bash":
		return "Hello from bash"
	case "verilog":
		return "Hello from verilog"
	case "lua":
		return "Hello from lua"
	default:
		return "Hello from py3"
	}
}

func sendRunRequest(t *testing.T, baseURL string, req model.RunRequest) model.RunResponse {
	t.Helper()

	statusCode, body := postRunExpectStatus(t, baseURL, req, http.StatusOK)
	require.Equal(t, http.StatusOK, statusCode, "unexpected HTTP status for /run: %s", string(body))

	var resp model.RunResponse
	require.NoError(t, json.Unmarshal(body, &resp), "unable to decode /run response: %s", string(body))
	return resp
}

func postRunExpectStatus(t *testing.T, baseURL string, req model.RunRequest, wantStatus int) (int, []byte) {
	t.Helper()

	body, err := json.Marshal(req)
	require.NoError(t, err)

	client := &http.Client{Timeout: 15 * time.Second}
	httpReq, err := http.NewRequest(http.MethodPost, baseURL+"/run", bytes.NewReader(body))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	rawBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, wantStatus, resp.StatusCode, "unexpected HTTP status for /run: %s", string(rawBody))

	return resp.StatusCode, rawBody
}

func serverURL(t *testing.T) string {
	t.Helper()

	if raw := strings.TrimSpace(os.Getenv("TEST_SERVER_URL")); raw != "" {
		return raw
	}
	return defaultServerURL
}

func requireServerReachable(t *testing.T, baseURL string) {
	t.Helper()

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/healthz")
	if err != nil || resp == nil || resp.StatusCode != http.StatusOK {
		t.Skipf("server not reachable at %s: %v", baseURL, err)
	}
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()

	cwd, err := os.Getwd()
	require.NoError(t, err)

	for dir := cwd; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "config", "languages.yaml")); err == nil {
			return dir
		}
		if dir == filepath.Dir(dir) {
			break
		}
	}

	return cwd
}
