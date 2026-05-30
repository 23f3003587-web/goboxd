package security

import (
	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/model"
)

// ValidateRunRequest performs full request validation
func ValidateRunRequest(req model.RunRequest) error {
	if err := config.ValidateRequest(req); err != nil {
		return err
	}
	if err := ValidateFilename(req.SourceFilename); err != nil {
		return err
	}
	if err := ValidateFilename(req.ArtifactFilename); err != nil {
		return err
	}
	// TODO: Flag validation (Hole 3) - expand later
	return nil
}
