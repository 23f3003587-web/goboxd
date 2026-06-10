package languages

// Schema documents the expected language configuration fields used by the service.
type Schema struct {
	ID        string `yaml:"id"`
	Name      string `yaml:"name"`
	BuildCmd  string `yaml:"build_cmd,omitempty"`
	RunCmd    string `yaml:"run_cmd"`
	TimeoutMs int    `yaml:"timeout_ms,omitempty"`
}
