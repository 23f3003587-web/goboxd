package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Language struct {
	ID             string `yaml:"id"`
	Name           string `yaml:"name"`
	SourceFilename string `yaml:"source_filename,omitempty"`
	Artifact       string `yaml:"artifact,omitempty"`

	// Strategy fields for languages like Java
	SourceFilenameStrategy   string `yaml:"source_filename_strategy,omitempty"`
	ArtifactFilenameStrategy string `yaml:"artifact_filename_strategy,omitempty"`

	Build *Step `yaml:"build,omitempty"`
	Run   Step  `yaml:"run"`
}

type Step struct {
	Cmd           string   `yaml:"cmd"`
	Args          []string `yaml:"args,omitempty"`
	Limits        Limits   `yaml:"limits"`
	FlagAllowlist []string `yaml:"flag_allowlist,omitempty"`
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

func LoadLanguages(path string) error {
	// Path fallback logic (keep your existing)
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
		return fmt.Errorf("invalid languages.yaml: %w", err)
	}

	if len(cfg.Languages) == 0 {
		return fmt.Errorf("no languages defined in yaml")
	}

	Languages = make(map[string]Language) // reset
	for _, lang := range cfg.Languages {
		if lang.ID == "" || lang.Run.Cmd == "" {
			return fmt.Errorf("language %q missing required id or run.cmd", lang.Name)
		}
		Languages[lang.ID] = lang
	}

	return nil
}

func GetLanguage(id string) (Language, bool) {
	lang, ok := Languages[id]
	return lang, ok
}
