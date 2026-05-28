package model

type RunRequest struct {
	Language         string     `json:"language"`
	Source           string     `json:"source"`
	SourceFilename   string     `json:"source_filename,omitempty"`
	ArtifactFilename string     `json:"artifact_filename,omitempty"`
	Build            *BuildRun  `json:"build,omitempty"`
	Run              *BuildRun  `json:"run,omitempty"`
	Tests            []TestCase `json:"tests"`
}

type BuildRun struct {
	Limits Limits   `json:"limits,omitempty"`
	Flags  []string `json:"flags,omitempty"`
}

type Limits struct {
	WallTimeS   int `json:"wall_time_s"`
	MemoryKB    int `json:"memory_kb"`
	MaxProcesses int `json:"max_processes"`
}

type TestCase struct {
	Stdin          string `json:"stdin"`
	ExpectedStdout string `json:"expected_stdout"`
}
