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
	default:
		return nil, fmt.Errorf("unknown rule: %s", ruleName)
	}
}

type NoCyrillicRule struct{}

func (r *NoCyrillicRule) Validate(text string) error {
	if regexp.MustCompile(`[\p{Cyrillic}]`).MatchString(text) {
		return fmt.Errorf("text contains Cyrillic characters")
	}
	return nil
}

type NoLatinRule struct{}

func (r *NoLatinRule) Validate(text string) error {
	if regexp.MustCompile(`[a-zA-Z]`).MatchString(text) {
		return fmt.Errorf("text contains Latin characters")
	}
	return nil
}

type NoDigitsRule struct{}

func (r *NoDigitsRule) Validate(text string) error {
	if regexp.MustCompile(`\d`).MatchString(text) {
		return fmt.Errorf("text contains digits")
	}
	return nil
}

type CyrillicOnlyRule struct{}

func (r *CyrillicOnlyRule) Validate(text string) error {
	if regexp.MustCompile(`[^\p{Cyrillic}\s\p{P}]`).MatchString(text) {
		return fmt.Errorf("text contains non-Cyrillic characters")
	}
	return nil
}

type LatinOnlyRule struct{}

func (r *LatinOnlyRule) Validate(text string) error {
	if regexp.MustCompile(`[^a-zA-Z\s\p{P}]`).MatchString(text) {
		return fmt.Errorf("text contains non-Latin characters")
	}
	return nil
}

type DigitsOnlyRule struct{}

func (r *DigitsOnlyRule) Validate(text string) error {
	if regexp.MustCompile(`[^\d\s\p{P}]`).MatchString(text) {
		return fmt.Errorf("text contains non-digit characters")
	}
	return nil
}
