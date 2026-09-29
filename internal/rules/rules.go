package rules

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/AnruKitakaze/commit-msg-guardian/config"
)

var (
	scopeSegment = regexp.MustCompile(`^[a-zA-Z0-9]+(?:[a-zA-Z0-9-]*[a-zA-Z0-9])?$`)
	scopePath    = regexp.MustCompile(`^[a-zA-Z0-9]+(?:[a-zA-Z0-9-]*[a-zA-Z0-9])?(?:/[a-zA-Z0-9]+(?:[a-zA-Z0-9-]*[a-zA-Z0-9])?)*$`)
)

// Checker applies one field's declarative text policy.
type Checker struct {
	allowed, required, forbidden        map[string]bool
	caseMode, trailingPeriod, structure string
	maxLines                            int
}

func Compile(policy config.TextPolicy) (Checker, error) {
	checker := Checker{
		caseMode:       policy.Case,
		trailingPeriod: policy.TrailingPeriod,
		structure:      policy.Structure,
		maxLines:       policy.MaxLines,
	}
	var err error
	if checker.allowed, err = compileGroups(policy.Characters.Allowed); err != nil {
		return Checker{}, fmt.Errorf("characters.allowed: %w", err)
	}
	if policy.Characters.Allowed != nil && len(checker.allowed) == 0 {
		return Checker{}, fmt.Errorf("characters.allowed must not be empty; omit it to allow all groups")
	}
	if checker.required, err = compileGroups(policy.Characters.Required); err != nil {
		return Checker{}, fmt.Errorf("characters.required: %w", err)
	}
	if checker.forbidden, err = compileGroups(policy.Characters.Forbidden); err != nil {
		return Checker{}, fmt.Errorf("characters.forbidden: %w", err)
	}
	for _, group := range sortedGroups(checker.required) {
		if groupIsForbidden(group, checker.forbidden) {
			return Checker{}, fmt.Errorf("characters.required group %q is forbidden by characters.forbidden", group)
		}
		if checker.allowed != nil && !requiredCanBeAllowed(group, checker.allowed) {
			return Checker{}, fmt.Errorf("characters.required group %q cannot occur under characters.allowed", group)
		}
	}
	for _, group := range sortedGroups(checker.allowed) {
		if groupIsForbidden(group, checker.forbidden) {
			return Checker{}, fmt.Errorf("characters.allowed group %q is completely forbidden by characters.forbidden", group)
		}
	}
	if checker.caseMode != "" && checker.caseMode != "lower-first-letter" && checker.caseMode != "upper-first-letter" {
		return Checker{}, fmt.Errorf("case must be lower-first-letter or upper-first-letter")
	}
	if checker.trailingPeriod != "" && checker.trailingPeriod != "require" && checker.trailingPeriod != "forbid" {
		return Checker{}, fmt.Errorf("trailing_period must be require or forbid")
	}
	if checker.structure != "" && checker.structure != "flat" && checker.structure != "slash-separated" {
		return Checker{}, fmt.Errorf("structure must be flat or slash-separated")
	}
	if checker.maxLines < 0 {
		return Checker{}, fmt.Errorf("max_lines must be non-negative")
	}
	if checker.trailingPeriod == "require" {
		if checker.structure != "" {
			return Checker{}, fmt.Errorf("trailing_period: require conflicts with structure: %s", checker.structure)
		}
		if !checker.acceptsRune('.') {
			return Checker{}, fmt.Errorf("trailing_period: require conflicts with characters.allowed or characters.forbidden: period is not permitted")
		}
	}
	if checker.structure != "" {
		alphabet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-"
		if checker.structure == "slash-separated" {
			alphabet += "/"
		}
		if err := checker.CheckAlphabet(checker.structure+" structure", alphabet); err != nil {
			return Checker{}, err
		}
	}
	return checker, nil
}

func sortedGroups(groups map[string]bool) []string {
	result := make([]string, 0, len(groups))
	for group := range groups {
		result = append(result, group)
	}
	sort.Strings(result)
	return result
}

func groupIsForbidden(group string, forbidden map[string]bool) bool {
	return forbidden[group] || isScriptGroup(group) && forbidden["letter"]
}

func compileGroups(values []string) (map[string]bool, error) {
	if values == nil {
		return nil, nil
	}
	result := make(map[string]bool, len(values))
	for _, value := range values {
		group := strings.ToLower(strings.ReplaceAll(value, "-", "_"))
		if !validGroup(group) {
			return nil, fmt.Errorf("unknown character group %q", value)
		}
		if result[group] {
			return nil, fmt.Errorf("duplicate character group %q", value)
		}
		result[group] = true
	}
	return result, nil
}

func validGroup(group string) bool {
	switch group {
	case "letter", "other_letter", "digit", "whitespace", "punctuation", "symbol", "mark", "other":
		return true
	case "common", "inherited":
		return false
	}
	for name := range unicode.Scripts {
		if strings.ToLower(name) == group {
			return true
		}
	}
	return false
}

func isScriptGroup(group string) bool {
	return group == "other_letter" || validGroup(group) && group != "letter" && group != "digit" &&
		group != "whitespace" && group != "punctuation" && group != "symbol" && group != "mark" && group != "other"
}

func requiredCanBeAllowed(group string, allowed map[string]bool) bool {
	if allowed[group] || isScriptGroup(group) && allowed["letter"] {
		return true
	}
	if group == "letter" {
		for allowedGroup := range allowed {
			if isScriptGroup(allowedGroup) {
				return true
			}
		}
	}
	return false
}

func (checker Checker) acceptsRune(char rune) bool {
	group, broader := groupsFor(char)
	return (checker.allowed == nil || checker.allowed[group] || checker.allowed[broader]) &&
		!checker.forbidden[group] && !checker.forbidden[broader]
}

// CheckAlphabet rejects field policies that cannot match the characters
// accepted by a fixed header grammar, such as a Conventional type or scope.
func (checker Checker) CheckAlphabet(field, alphabet string) error {
	possible := make(map[string]bool)
	hasScopeStart := false
	for _, char := range alphabet {
		if !checker.acceptsRune(char) {
			continue
		}
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' {
			hasScopeStart = true
		}
		group, broader := groupsFor(char)
		possible[group] = true
		if broader != "" {
			possible[broader] = true
		}
	}
	if len(possible) == 0 {
		return fmt.Errorf("characters policy permits no characters in %s", field)
	}
	if checker.structure != "" && !hasScopeStart {
		return fmt.Errorf("structure: %s needs a Latin letter or digit in %s, but characters policy excludes them", checker.structure, field)
	}
	for _, group := range sortedGroups(checker.required) {
		if !possible[group] {
			return fmt.Errorf("characters.required group %q cannot occur in %s", group, field)
		}
	}
	if checker.trailingPeriod == "require" && !strings.ContainsRune(alphabet, '.') {
		return fmt.Errorf("trailing_period: require cannot match %s", field)
	}
	return nil
}

func (checker Checker) Validate(value string) error {
	if checker.allowed == nil && len(checker.required) == 0 && len(checker.forbidden) == 0 &&
		checker.caseMode == "" && checker.trailingPeriod == "" && checker.structure == "" && checker.maxLines == 0 {
		return nil
	}
	seen := make(map[string]bool)
	for _, char := range value {
		group, broader := groupsFor(char)
		seen[group] = true
		if broader != "" {
			seen[broader] = true
		}
		if checker.allowed != nil && !checker.allowed[group] && !checker.allowed[broader] {
			return fmt.Errorf("character %q is not allowed by characters.allowed", char)
		}
		if checker.forbidden[group] || checker.forbidden[broader] {
			return fmt.Errorf("character %q is forbidden by characters.forbidden", char)
		}
	}
	for _, group := range sortedGroups(checker.required) {
		if !seen[group] {
			return fmt.Errorf("required character group %q is missing", group)
		}
	}
	if checker.structure == "flat" && !scopeSegment.MatchString(value) {
		return fmt.Errorf("text must be one scope segment with Latin letters, digits, or hyphens")
	}
	if checker.structure == "slash-separated" && !scopePath.MatchString(value) {
		return fmt.Errorf("text must be slash-delimited scope segments with Latin letters, digits, or hyphens")
	}
	for _, char := range value {
		if !unicode.IsLetter(char) {
			continue
		}
		if checker.caseMode == "lower-first-letter" && !unicode.IsLower(char) || checker.caseMode == "upper-first-letter" && !unicode.IsUpper(char) {
			return fmt.Errorf("text must start with a %s letter", strings.TrimSuffix(checker.caseMode, "-first-letter"))
		}
		break
	}
	if checker.trailingPeriod == "require" && !strings.HasSuffix(value, ".") {
		return fmt.Errorf("text must end with a period")
	}
	if checker.trailingPeriod == "forbid" && strings.HasSuffix(value, ".") {
		return fmt.Errorf("text must not end with a period")
	}
	if checker.maxLines > 0 && strings.Count(value, "\n")+1 > checker.maxLines {
		return fmt.Errorf("text must contain at most %d line(s)", checker.maxLines)
	}
	return nil
}

func groupsFor(char rune) (specific, broader string) {
	if unicode.IsLetter(char) {
		for name, table := range unicode.Scripts {
			if name != "Common" && name != "Inherited" && unicode.Is(table, char) {
				return strings.ToLower(name), "letter"
			}
		}
		return "other_letter", "letter"
	}
	switch {
	case unicode.IsDigit(char):
		return "digit", ""
	case unicode.IsSpace(char):
		return "whitespace", ""
	case unicode.IsPunct(char):
		return "punctuation", ""
	case unicode.IsSymbol(char):
		return "symbol", ""
	case unicode.IsMark(char):
		return "mark", ""
	default:
		return "other", ""
	}
}
