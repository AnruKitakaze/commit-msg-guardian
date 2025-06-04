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
	typeRules := flag.String("type-rules", "latinOnly", "Comma-separated rules for commit type")
	scopeRules := flag.String("scope-rules", "latinOnly,digitsOnly", "Comma-separated rules for commit scope")
	descriptionRules := flag.String("description-rules", "noCyrillic", "Comma-separated rules for commit description")

	flag.Parse()

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
	typeRulesList := strings.Split(*typeRules, ",")
	scopeRulesList := strings.Split(*scopeRules, ",")
	descriptionRulesList := strings.Split(*descriptionRules, ",")

	// Validate commit message
	if err := msg.ValidateWithRules(typeRulesList, scopeRulesList, descriptionRulesList); err != nil {
		fmt.Fprintf(os.Stderr, "Commit message validation failed: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
