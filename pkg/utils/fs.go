package utils

import "os"

// EnsureDir creates a directory if it does not already exist.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}
