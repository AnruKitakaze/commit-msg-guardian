package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/AnruKitakaze/commit-msg-guardian/parser"
)

func main() {
	// Define flags for rules
	typeRules := flag.String("type-rules", "allowLatin", "Comma-separated rules for commit type")
	scopeRules := flag.String("scope-rules", "allowScope", "Comma-separated rules for commit scope")
	descriptionRules := flag.String("description-rules", "noCyrillic", "Comma-separated rules for commit description")
	bodyRules := flag.String("body-rules", "", "Comma-separated rules for commit body")
	descriptionLengthLimit := flag.Int("description-length-limit", 0, "Maximum allowed description length; 0 disables the limit")
	bodyLengthLimit := flag.Int("body-length-limit", 0, "Maximum allowed body length; 0 disables the limit")

	flag.Parse()

	if *descriptionLengthLimit < 0 {
		fmt.Fprintln(os.Stderr, "Error: description length limit must be non-negative")
		os.Exit(1)
	}
	if *bodyLengthLimit < 0 {
		fmt.Fprintln(os.Stderr, "Error: body length limit must be non-negative")
		os.Exit(1)
	}

	// Get commit message file path from arguments
	if len(flag.Args()) < 1 {
		fmt.Fprintln(os.Stderr, "Error: commit message file path is required")
		os.Exit(1)
	}

	// Read commit message file
	commitMsgFile := flag.Args()[0]
	commitMsg, err := os.ReadFile(commitMsgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading commit message file: %v\n", err)
		os.Exit(1)
	}

	// Parse commit message
	msg, err := parser.ParseCommitMessage(string(commitMsg))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing commit message: %v\n", err)
		os.Exit(1)
	}

	// Split rules into slices
	typeRulesList := splitRules(*typeRules)
	scopeRulesList := splitRules(*scopeRules)
	descriptionRulesList := splitRules(*descriptionRules)
	bodyRulesList := splitRules(*bodyRules)

	// Validate commit message
	if err := msg.ValidateWithRules(typeRulesList, scopeRulesList, descriptionRulesList, bodyRulesList); err != nil {
		fmt.Fprintf(os.Stderr, "Commit message validation failed: %v\n", err)
		os.Exit(1)
	}
	if err := msg.ValidateLengthLimits(*descriptionLengthLimit, *bodyLengthLimit); err != nil {
		fmt.Fprintf(os.Stderr, "Commit message validation failed: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func splitRules(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	ruleNames := strings.Split(value, ",")
	rules := make([]string, 0, len(ruleNames))
	for _, ruleName := range ruleNames {
		ruleName = strings.TrimSpace(ruleName)
		if ruleName == "" {
			continue
		}
		rules = append(rules, ruleName)
	}
	return rules
}
