package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Language struct {
	ID             string `yaml:"id"`
	Name           string `yaml:"name"`
	SourceFilename string `yaml:"source_filename"`
	Run            struct {
		Cmd    string   `yaml:"cmd"`
		Args   []string `yaml:"args"`
		Limits Limits   `yaml:"limits"`
	} `yaml:"run"`
}

type Limits struct {
	WallTimeS   int `yaml:"wall_time_s"`
	MemoryKB    int `yaml:"memory_kb"`
	MaxProcesses int `yaml:"max_processes"`
}

type Config struct {
	Languages []Language `yaml:"languages"`
}

var Languages = make(map[string]Language)

func LoadLanguages(path string) error {
	data, err := os.ReadFile(path)
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

func GetLanguage(id string) (Language, bool) {
	lang, ok := Languages[id]
	return lang, ok
}
