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

	lang, ok := config.GetLanguage(req.Language)
	if !ok {
		return fmt.Errorf("unknown language: %s", req.Language)
	}

	// Validate build limits and flags
	if req.Build != nil {
		if lang.Build == nil {
			return fmt.Errorf("language %q does not support build overrides", req.Language)
		}
		if err := validateLimits(req.Build.Limits, lang.Build.Limits, "build"); err != nil {
			return err
		}
		if err := ValidateBuildFlags(req.Language, req.Build.Flags); err != nil {
			return err
		}
	}

	// Validate run limits and flags
	if req.Run != nil {
		if err := validateLimits(req.Run.Limits, lang.Run.Limits, "run"); err != nil {
			return err
		}
		if err := ValidateRunFlags(req.Language, req.Run.Flags); err != nil {
			return err
		}
	}

	return nil
}

// validateLimits ensures user-supplied limits do not exceed the language's configured maximums
func validateLimits(user model.Limits, langLimits config.Limits, step string) error {
	if user.WallTimeS > langLimits.WallTimeS {
		return fmt.Errorf("%s: wall_time_s %d exceeds maximum allowed %d", step, user.WallTimeS, langLimits.WallTimeS)
	}
	if user.MemoryKB > langLimits.MemoryKB {
		return fmt.Errorf("%s: memory_kb %d exceeds maximum allowed %d", step, user.MemoryKB, langLimits.MemoryKB)
	}
	if user.MaxProcesses > langLimits.MaxProcesses {
		return fmt.Errorf("%s: max_processes %d exceeds maximum allowed %d", step, user.MaxProcesses, langLimits.MaxProcesses)
	}
	return nil
}
