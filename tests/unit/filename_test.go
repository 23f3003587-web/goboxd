package unit_test

import (
	"strings"
	"testing"

	"github.com/thesouldev/goboxd/internal/security"
)

func TestValidateFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{"empty", "", false},
		{"valid python", "solution.py", false},
		{"valid cpp", "main.cpp", false},
		{"valid with dot", "file.txt", false},
		{"traversal with ..", "../../etc/passwd", true},
		{"absolute path", "/etc/passwd", true},
		{"hidden file", ".hidden", true},
		{"too long", strings.Repeat("a", 101), true},
		{"contains slash", "dir/file.py", true},
		{"contains backslash", "dir\\file.py", true},
		{"contains null byte", "file\x00.py", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := security.ValidateFilename(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilename(%q) error = %v, wantErr %v", tt.filename, err, tt.wantErr)
			}
		})
	}
}
