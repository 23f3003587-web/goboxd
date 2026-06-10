package utils

import "os"

// TempDir creates a temporary working directory for sandbox runs.
func TempDir(dir string) (string, error) {
	return os.MkdirTemp(dir, "goboxd-")
}
