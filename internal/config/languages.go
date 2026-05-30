package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Language struct {
	ID             string `yaml:"id"`
	Name           string `yaml:"name"`
	SourceFilename string `yaml:"source_filename"`
	Artifact       string `yaml:"artifact,omitempty"` // Added for compiled languages like C++

	Build *Step `yaml:"build,omitempty"` // Pointer used because build is optional (e.g., Python doesn't have it)
	Run   Step  `yaml:"run"`
}

type Step struct {
	Cmd           string   `yaml:"cmd"`
	Args          []string `yaml:"args,omitempty"`
	Limits        Limits   `yaml:"limits"`
	FlagAllowlist []string `yaml:"flag_allowlist,omitempty"` // Added to support C++ flags restriction
}

type Limits struct {
	WallTimeS    int `yaml:"wall_time_s"`
	MemoryKB     int `yaml:"memory_kb"`
	MaxProcesses int `yaml:"max_processes"`
}

type Config struct {
	Languages []Language `yaml:"languages"`
}

var Languages = make(map[string]Language)

// LoadLanguages reads the YAML file and populates the global Languages map.
func LoadLanguages(path string) error {
	// Try the provided path first, then walk up a few parent dirs to support
	// running tests from package subdirectories.
	var data []byte
	var err error
	tryPath := path
	for i := 0; i < 4; i++ {
		data, err = os.ReadFile(tryPath)
		if err == nil {
			break
		}
		if !os.IsNotExist(err) {
			return err
		}
		tryPath = filepath.Join("..", tryPath)
	}
	if err != nil {
		return err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}
	for _, lang := range cfg.Languages {
		Languages[lang.ID] = lang
	}
	return nil
}

// GetLanguage retrieves a language configuration by its ID string.
func GetLanguage(id string) (Language, bool) {
	lang, ok := Languages[id]
	return lang, ok
}