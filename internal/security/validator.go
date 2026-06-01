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
	// Flag validation (Hole 3) - validates against language's flag_allowlist
	if req.Build != nil && len(req.Build.Flags) > 0 {
		if err := ValidateBuildFlags(req.Language, req.Build.Flags); err != nil {
			return err
		}
	}
	if req.Run != nil && len(req.Run.Flags) > 0 {
		if err := ValidateRunFlags(req.Language, req.Run.Flags); err != nil {
			return err
		}
	}
	return nil
}
