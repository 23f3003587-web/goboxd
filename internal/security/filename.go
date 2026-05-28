package security

import (
	"fmt"
	"strings"
)

func ValidateFilename(name string) error {
	if name == "" {
		return nil
	}
	if len(name) > 100 {
		return fmt.Errorf("filename too long")
	}
	if strings.ContainsAny(name, "/\\") || strings.Contains(name, "..") {
		return fmt.Errorf("invalid filename: directory traversal detected")
	}
	if strings.HasPrefix(name, ".") || strings.Contains(name, "\x00") {
		return fmt.Errorf("invalid filename: hidden or malformed")
	}
	return nil
}
