package config

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the versioned, project-owned validation policy.
type Config struct {
	Version int          `yaml:"version"`
	Commit  CommitConfig `yaml:"commit"`
	CI      CIConfig     `yaml:"ci"`
}

type CommitConfig struct {
	Formats []Format `yaml:"formats"`
}

type Format struct {
	ID         string   `yaml:"id"`
	Pattern    string   `yaml:"pattern"`
	Separators []string `yaml:"separators"`
	When       *When    `yaml:"when"`
	Type       *Type    `yaml:"type"`
	Scope      *Scope   `yaml:"scope"`
	Subject    Text     `yaml:"subject"`
	Body       Body     `yaml:"body"`
	Trailers   Trailers `yaml:"trailers"`
}

type When struct {
	AuthorEmail string `yaml:"author_email"`
}

type Type struct {
	Allowed    []string `yaml:"allowed"`
	TextPolicy `yaml:",inline"`
}

type Scope struct {
	Required   *bool    `yaml:"required"`
	Brackets   string   `yaml:"brackets"`
	Allowed    []string `yaml:"allowed"`
	Issue      *Issue   `yaml:"issue"`
	TextPolicy `yaml:",inline"`
}

type Issue struct {
	Projects []string `yaml:"projects"`
}

type Text struct {
	MinLength  int `yaml:"min_length"`
	MaxLength  int `yaml:"max_length"`
	TextPolicy `yaml:",inline"`
}

type Body struct {
	Required         bool     `yaml:"required"`
	OptionalForTypes []string `yaml:"optional_for_types"`
	MinLength        int      `yaml:"min_length"`
	MaxLineLength    int      `yaml:"max_line_length"`
	MaxTotalLength   int      `yaml:"max_total_length"`
	TextPolicy       `yaml:",inline"`
}

type TextPolicy struct {
	Characters     Characters `yaml:"characters"`
	Structure      string     `yaml:"structure"`
	Case           string     `yaml:"case"`
	TrailingPeriod string     `yaml:"trailing_period"`
	MaxLines       int        `yaml:"max_lines"`
}

type Characters struct {
	Allowed   []string `yaml:"allowed"`
	Required  []string `yaml:"required"`
	Forbidden []string `yaml:"forbidden"`
}

type Trailers struct {
	Required []RequiredTrailer `yaml:"required"`
}

type RequiredTrailer struct {
	Name         string `yaml:"name"`
	ValuePattern string `yaml:"value_pattern"`
}

type CIConfig struct {
	Commits CICommits `yaml:"commits"`
}

type CICommits struct {
	Base CIBase `yaml:"base"`
}

type CIBase struct {
	Ref string `yaml:"ref"`
	Env string `yaml:"env"`
}

// Load rejects unknown keys and extra YAML documents before validation.
func Load(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config %q: %w", path, err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)
	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		if err == io.EOF {
			return Config{}, fmt.Errorf("config %q is empty; version and commit.formats are required", path)
		}
		return Config{}, fmt.Errorf("parse config %q: %w", path, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Config{}, fmt.Errorf("config %q must contain exactly one YAML document", path)
		}
		return Config{}, fmt.Errorf("parse config %q: %w", path, err)
	}
	return cfg, nil
}
