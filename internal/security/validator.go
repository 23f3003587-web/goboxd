package security

import (
	"fmt"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/model"
)

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

	if _, ok := config.GetLanguage(req.Language); !ok {
		return fmt.Errorf("unknown language: %s", req.Language)
	}

	// Validate build and run flags using the language-specific rules.
	if req.Build != nil {
		if err := ValidateBuildFlags(req.Language, req.Build.Flags); err != nil {
			return err
		}
	}
	if req.Run != nil {
		if err := ValidateRunFlags(req.Language, req.Run.Flags); err != nil {
			return err
		}
	}

	return nil
}
