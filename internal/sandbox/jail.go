package sandbox

import (
	"os"
	"path/filepath"

	"github.com/thesouldev/goboxd/internal/security"
)

type Jail struct {
	Path string
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
	fullPath := filepath.Join(j.Path, filename)
	return os.WriteFile(fullPath, []byte(source), 0644)
}

func (j *Jail) Cleanup() {
	os.RemoveAll(j.Path)
}
