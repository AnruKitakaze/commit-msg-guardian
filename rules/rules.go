package rules

import (
	"fmt"
	"regexp"
	"strings"
)

// ConventionalCommitTypes defines allowed commit types
var ConventionalCommitTypes = []string{
	"feat", "fix", "docs", "style", "refactor",
	"perf", "test", "build", "ci", "chore", "revert",
}

// Rule represents a validation rule
type Rule interface {
	Validate(text string) error
}

// RuleFactory creates a Rule based on the rule name
func RuleFactory(ruleName string) (Rule, error) {
	switch strings.ToLower(ruleName) {
	case "nocyrillic":
		return &NoCyrillicRule{}, nil
	case "nolatin":
		return &NoLatinRule{}, nil
	case "nodigits":
		return &NoDigitsRule{}, nil
	case "cyrilliconly":
		return &CyrillicOnlyRule{}, nil
	case "latinonly":
		return &LatinOnlyRule{}, nil
	case "digitsonly":
		return &DigitsOnlyRule{}, nil
	case "allowlatin":
		return &AllowLatinRule{}, nil
	case "allowcyrillic":
		return &AllowCyrillicRule{}, nil
	case "allowdigits":
		return &AllowDigitsRule{}, nil
	case "allowscope":
		return &AllowScopeRule{}, nil
	default:
		return nil, fmt.Errorf("unknown rule: %s", ruleName)
	}
}

// NoCyrillicRule prevents Cyrillic characters
type NoCyrillicRule struct{}

func (r *NoCyrillicRule) Validate(text string) error {
	if regexp.MustCompile(`[\p{Cyrillic}]`).MatchString(text) {
		return fmt.Errorf("text contains Cyrillic characters")
	}
	return nil
}

// NoLatinRule prevents Latin characters
type NoLatinRule struct{}

func (r *NoLatinRule) Validate(text string) error {
	if regexp.MustCompile(`[a-zA-Z]`).MatchString(text) {
		return fmt.Errorf("text contains Latin characters")
	}
	return nil
}

// NoDigitsRule prevents digits
type NoDigitsRule struct{}

func (r *NoDigitsRule) Validate(text string) error {
	if regexp.MustCompile(`\d`).MatchString(text) {
		return fmt.Errorf("text contains digits")
	}
	return nil
}

// CyrillicOnlyRule allows only Cyrillic characters and basic punctuation
type CyrillicOnlyRule struct{}

func (r *CyrillicOnlyRule) Validate(text string) error {
	if regexp.MustCompile(`[^\p{Cyrillic}\s\p{P}]`).MatchString(text) {
		return fmt.Errorf("text contains non-Cyrillic characters")
	}
	return nil
}

// LatinOnlyRule allows only Latin characters and basic punctuation
type LatinOnlyRule struct{}

func (r *LatinOnlyRule) Validate(text string) error {
	if regexp.MustCompile(`[^a-zA-Z\s\p{P}]`).MatchString(text) {
		return fmt.Errorf("text contains non-Latin characters")
	}
	return nil
}

// DigitsOnlyRule allows only digits and basic punctuation
type DigitsOnlyRule struct{}

func (r *DigitsOnlyRule) Validate(text string) error {
	if regexp.MustCompile(`[^\d\s\p{P}]`).MatchString(text) {
		return fmt.Errorf("text contains non-digit characters")
	}
	return nil
}

// AllowLatinRule allows Latin characters but doesn't require them
type AllowLatinRule struct{}

func (r *AllowLatinRule) Validate(text string) error {
	if regexp.MustCompile(`[^a-zA-Z\s\p{P}\d-]`).MatchString(text) {
		return fmt.Errorf("text contains characters that are not Latin, digits, spaces, or punctuation")
	}
	return nil
}

// AllowCyrillicRule allows Cyrillic characters but doesn't require them
type AllowCyrillicRule struct{}

func (r *AllowCyrillicRule) Validate(text string) error {
	if regexp.MustCompile(`[^\p{Cyrillic}\s\p{P}\d-]`).MatchString(text) {
		return fmt.Errorf("text contains characters that are not Cyrillic, digits, spaces, or punctuation")
	}
	return nil
}

// AllowDigitsRule allows digits but doesn't require them
type AllowDigitsRule struct{}

func (r *AllowDigitsRule) Validate(text string) error {
	if regexp.MustCompile(`[^a-zA-Z\d\s\p{P}-]`).MatchString(text) {
		return fmt.Errorf("text contains characters that are not Latin, digits, spaces, or punctuation")
	}
	return nil
}

// AllowScopeRule allows characters valid in scope (Latin, digits, and hyphens)
type AllowScopeRule struct{}

func (r *AllowScopeRule) Validate(text string) error {
	if !regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]$`).MatchString(text) {
		return fmt.Errorf("scope must start and end with alphanumeric character and contain only Latin letters, digits, and hyphens")
	}
	return nil
}
