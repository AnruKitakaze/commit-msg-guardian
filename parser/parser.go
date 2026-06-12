package parser

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/AnruKitakaze/commit-msg-guardian/rules"
)

// CommitMessage represents a parsed commit message
type CommitMessage struct {
	Type           string
	Scope          string
	BreakingChange bool
	Description    string
	Body           string
}

// ParseCommitMessage parses a commit message into its components
func ParseCommitMessage(message string) (*CommitMessage, error) {
	lines := strings.SplitN(message, "\n", 2)
	header := lines[0]

	// Parse header
	headerPattern := regexp.MustCompile(`^(\w+)(?:\(([\w-]+(?:/[\w-]+)*)\))?(!)?: (.+)$`)
	matches := headerPattern.FindStringSubmatch(header)
	if matches == nil {
		return nil, fmt.Errorf("invalid commit message format")
	}

	commitType := matches[1]
	if !isValidCommitType(commitType) {
		return nil, fmt.Errorf("invalid commit type: %s", commitType)
	}

	body := ""
	if len(lines) > 1 {
		body = strings.TrimSpace(lines[1])
	}

	return &CommitMessage{
		Type:           commitType,
		Scope:          matches[2],
		BreakingChange: matches[3] == "!",
		Description:    matches[4],
		Body:           body,
	}, nil
}

func isValidCommitType(commitType string) bool {
	return slices.Contains(rules.ConventionalCommitTypes, commitType)
}

// ValidateWithRules validates different parts of the commit message with specified rules
func (cm *CommitMessage) ValidateWithRules(typeRules, scopeRules, descriptionRules, bodyRules []string) error {
	// Validate type
	if err := validateText(cm.Type, typeRules); err != nil {
		return fmt.Errorf("type validation failed: %w", err)
	}

	// Validate scope
	if cm.Scope != "" {
		if err := validateText(cm.Scope, scopeRules); err != nil {
			return fmt.Errorf("scope validation failed: %w", err)
		}
	}

	// Validate description
	if err := validateText(cm.Description, descriptionRules); err != nil {
		return fmt.Errorf("description validation failed: %w", err)
	}

	// Validate body
	if err := validateText(cm.Body, bodyRules); err != nil {
		return fmt.Errorf("body validation failed: %w", err)
	}

	return nil
}

// ValidateLengthLimits validates strict description and body length limits.
func (cm *CommitMessage) ValidateLengthLimits(descriptionLimit, bodyLimit int) error {
	if err := validateLengthLimit("description", cm.Description, descriptionLimit); err != nil {
		return err
	}
	if err := validateLengthLimit("body", cm.Body, bodyLimit); err != nil {
		return err
	}
	return nil
}

func validateLengthLimit(name, text string, limit int) error {
	if limit == 0 {
		return nil
	}
	length := len([]rune(text))
	if length > limit {
		return fmt.Errorf("%s must be no longer than %d characters, got %d", name, limit, length)
	}
	return nil
}

func validateText(text string, ruleNames []string) error {
	for _, ruleName := range ruleNames {
		rule, err := rules.RuleFactory(ruleName)
		if err != nil {
			return err
		}
		if err := rule.Validate(text); err != nil {
			return err
		}
	}
	return nil
}
