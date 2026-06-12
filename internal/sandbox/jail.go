package sandbox

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/thesouldev/goboxd/internal/security"
)

var (
	orphanCleanupOnce sync.Once
	orphanMaxAge      = 30 * time.Minute
)

type Jail struct {
	Path string
}

func init() {
	orphanCleanupOnce.Do(cleanupOrphanJails)
}

func cleanupOrphanJails() {
	paths, err := filepath.Glob("/tmp/goboxd-jail-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "warn: orphan jail cleanup failed: %v\n", err)
		return
	}
	cutoff := time.Now().Add(-orphanMaxAge)
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			continue
		}
		if info.ModTime().Before(cutoff) {
			if err := os.RemoveAll(path); err != nil {
				fmt.Fprintf(os.Stderr, "warn: failed to remove orphan jail %s: %v\n", path, err)
			}
		}
	}
}

func NewJail() (*Jail, error) {
	dir, err := os.MkdirTemp("/tmp", "goboxd-jail-*")
	if err != nil {
		return nil, err
	}
	return &Jail{Path: dir}, nil
}

func (j *Jail) WriteSource(filename, source string) error {
	if err := security.ValidateFilename(filename); err != nil {
		return err
	}
	fullPath := filepath.Clean(filepath.Join(j.Path, filename))
	rel, err := filepath.Rel(j.Path, fullPath)
	if err != nil || strings.HasPrefix(rel, "..") || rel == ".." {
		return fmt.Errorf("invalid filename: path escapes jail")
	}
	return os.WriteFile(fullPath, []byte(source), 0644)
}

func (j *Jail) Cleanup() {
	if err := os.RemoveAll(j.Path); err != nil {
		// log but don't return — cleanup should be best-effort
		fmt.Fprintf(os.Stderr, "warn: failed to remove jail %s: %v\n", j.Path, err)
	}
}
