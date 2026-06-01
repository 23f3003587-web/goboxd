package security

import (
	"fmt"
	"strings"

	"github.com/thesouldev/goboxd/internal/config"
)

// ValidateBuildFlags validates build flags against the language's build step allowlist
func ValidateBuildFlags(language string, flags []string) error {
	if len(flags) == 0 {
		return nil
	}

	lang, exists := config.GetLanguage(language)
	if !exists {
		return fmt.Errorf("unknown language: %s", language)
	}

	// If no build step, reject all build flags
	if lang.Build == nil {
		return fmt.Errorf("language '%s' does not support build phase", language)
	}

	// If no allowlist is defined, reject all flags
	if len(lang.Build.FlagAllowlist) == 0 {
		return fmt.Errorf("language '%s' does not allow custom build flags", language)
	}

	// Validate each flag against the allowlist
	for _, flag := range flags {
		if !isFlagAllowed(flag, lang.Build.FlagAllowlist) {
			return fmt.Errorf("build flag '%s' not in allowlist for language '%s'", flag, language)
		}
	}

	return nil
}

// ValidateRunFlags validates run flags against the language's run step allowlist
func ValidateRunFlags(language string, flags []string) error {
	if len(flags) == 0 {
		return nil
	}

	lang, exists := config.GetLanguage(language)
	if !exists {
		return fmt.Errorf("unknown language: %s", language)
	}

	// If no allowlist is defined, reject all flags
	if len(lang.Run.FlagAllowlist) == 0 {
		return fmt.Errorf("language '%s' does not allow custom run flags", language)
	}

	// Validate each flag against the allowlist
	for _, flag := range flags {
		if !isFlagAllowed(flag, lang.Run.FlagAllowlist) {
			return fmt.Errorf("run flag '%s' not in allowlist for language '%s'", flag, language)
		}
	}

	return nil
}

// isFlagAllowed checks if a flag matches any pattern in the allowlist
func isFlagAllowed(flag string, allowlist []string) bool {
	for _, allowed := range allowlist {
		if matchesPattern(flag, allowed) {
			return true
		}
	}
	return false
}

// matchesPattern checks if a flag matches an allowlist pattern
// Supports exact match and wildcard patterns (e.g., "-std=*")
func matchesPattern(flag, pattern string) bool {
	// Exact match
	if flag == pattern {
		return true
	}

	// Wildcard pattern matching (e.g., "-std=*" matches "-std=c++17")
	if strings.Contains(pattern, "*") {
		// Replace * with a regex-like match
		idx := strings.Index(pattern, "*")
		prefix := pattern[:idx]
		suffix := pattern[idx+1:]

		// Check if flag starts with prefix and ends with suffix
		if strings.HasPrefix(flag, prefix) && strings.HasSuffix(flag, suffix) {
			return true
		}
	}

	return false
}
