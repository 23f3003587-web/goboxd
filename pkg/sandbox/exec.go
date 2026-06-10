package sandbox

// ExecOptions carries sandbox execution settings for a language runner.
type ExecOptions struct {
	Language string
	Source   string
}

// ExecResult describes the outcome of a sandboxed run.
type ExecResult struct {
	Status string
	Output string
}
