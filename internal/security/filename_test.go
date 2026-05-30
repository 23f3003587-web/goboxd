package security

import (
	"strings"
	"testing"
)

func TestValidateFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{"empty", "", false},
		{"valid", "solution.py", false},
		{"valid cpp", "main.cpp", false},
		{"with dot", "file.txt", false},
		{"traversal", "../../etc/passwd", true},
		{"absolute", "/etc/passwd", true},
		{"hidden", ".hidden", true},
		{"too long", strings.Repeat("a", 101), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilename(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilename() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
