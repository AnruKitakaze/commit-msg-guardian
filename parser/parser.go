package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/AnruKitakaze/commit-msg-guardian/rules"
)

// CommitMessage represents a parsed commit message
type CommitMessage struct {
	Type        string
	Scope       string
	Description string
	Body        string
}

// ParseCommitMessage parses a commit message into its components
func ParseCommitMessage(message string) (*CommitMessage, error) {
	lines := strings.SplitN(message, "\n", 2)
	header := lines[0]

	// Parse header
	headerPattern := regexp.MustCompile(`^(\w+)(?:\(([\w-]+)\))?: (.+)$`)
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
		Type:        commitType,
		Scope:       matches[2],
		Description: matches[3],
		Body:        body,
	}, nil
}

func isValidCommitType(commitType string) bool {
	for _, validType := range rules.ConventionalCommitTypes {
		if commitType == validType {
			return true
		}
	}
	return false
}

// ValidateWithRules validates different parts of the commit message with specified rules
func (cm *CommitMessage) ValidateWithRules(typeRules, scopeRules, descriptionRules []string) error {
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
