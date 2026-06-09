package integration

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os/exec"
	"testing"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/handler"
)

func TestRunEndpoint_Stage1(t *testing.T) {
	// Stage 1 requirement: nsjail must be available
	if _, err := exec.LookPath("nsjail"); err != nil {
		t.Skip("nsjail not found in PATH; skipping integration tests")
	}

	// Load language configuration
	if err := config.LoadLanguages("languages.yaml"); err != nil {
		t.Fatalf("failed to load languages.yaml: %v", err)
	}

	tests := []struct {
		name       string
		req        string
		wantStatus int
	}{
		// ==================== Python 3 ====================
		{
			name: "py3 hello world",
			req: `{
				"language": "py3",
				"source": "print('Hello World')",
				"tests": [{"stdin": "", "expected_stdout": "Hello World"}]
			}`,
			wantStatus: 200,
		},
		{
			name: "py3 with input",
			req: `{
				"language": "py3",
				"source": "n = int(input())\nprint(n * 10)",
				"tests": [{"stdin": "8", "expected_stdout": "80"}]
			}`,
			wantStatus: 200,
		},

		// ==================== C++ ====================
		{
			name: "cpp hello world",
			req: `{
				"language": "cpp",
				"source": "#include <iostream>\nint main() { std::cout << \"Hello World\\n\"; return 0; }",
				"tests": [{"stdin": "", "expected_stdout": "Hello World"}]
			}`,
			wantStatus: 200,
		},
		{
			name: "cpp with input",
			req: `{
				"language": "cpp",
				"source": "#include <iostream>\nint main() { int a, b; std::cin >> a >> b; std::cout << (a + b) << std::endl; return 0; }",
				"tests": [{"stdin": "13 7", "expected_stdout": "20"}]
			}`,
			wantStatus: 200,
		},
		{
			name: "cpp with allowed build flags",
			req: `{
				"language": "cpp",
				"source": "#include <iostream>\nint main() { std::cout << \"Optimized\\n\"; return 0; }",
				"build": {"flags": ["-O2", "-Wall"]},
				"tests": [{"stdin": "", "expected_stdout": "Optimized"}]
			}`,
			wantStatus: 200,
		},

		// ==================== Security & Validation ====================
		{
			name: "path traversal blocked",
			req: `{
				"language": "cpp",
				"source_filename": "../../malicious.cpp",
				"source": "#include <iostream>\nint main(){return 0;}",
				"tests": [{"stdin": "", "expected_stdout": ""}]
			}`,
			wantStatus: 400,
		},
		{
			name: "disallowed build flag",
			req: `{
				"language": "cpp",
				"source": "#include <iostream>\nint main(){return 0;}",
				"build": {"flags": ["-fpermissive", "-O999"]},
				"tests": [{"stdin": "", "expected_stdout": ""}]
			}`,
			wantStatus: 400,
		},
		{
			name: "unknown language",
			req: `{
				"language": "rust",
				"source": "fn main() {}",
				"tests": [{"stdin": "", "expected_stdout": ""}]
			}`,
			wantStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/run", bytes.NewBufferString(tt.req))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.Run(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("%s: wrong status code\ngot: %d\nwant: %d\nbody: %s",
					tt.name, rr.Code, tt.wantStatus, rr.Body.String())
			}

			// Additional validation for successful responses
			if tt.wantStatus == 200 {
				var resp map[string]any
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("%s: successful response must be valid JSON", tt.name)
				}
				if _, hasStatus := resp["status"]; !hasStatus {
					t.Errorf("%s: response should contain 'status' field", tt.name)
				}
			}
		})
	}
}
