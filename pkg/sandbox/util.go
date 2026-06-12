package sandbox

// Truncate returns the input string, preserving the caller-provided limit when used by the service.
func Truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max]
}
