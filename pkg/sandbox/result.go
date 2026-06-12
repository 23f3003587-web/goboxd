package sandbox

// Result wraps the exit status and captured output from an execution attempt.
type Result struct {
	ExitCode int
	Stdout   string
	Stderr   string
}
