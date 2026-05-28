package model

type RunResponse struct {
	Status string       `json:"status"`
	Build  BuildResult  `json:"build"`
	Tests  []TestResult `json:"tests"`
}

type BuildResult struct {
	Status     string `json:"status"`
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	DurationMS int    `json:"duration_ms"`
}

type TestResult struct {
	Status       string `json:"status"`
	Stdout       string `json:"stdout"`
	Stderr       string `json:"stderr"`
	DurationMS   int    `json:"duration_ms"`
	MemoryPeakKB int    `json:"memory_peak_kb"`
}
