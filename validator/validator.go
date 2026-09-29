package validator

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/AnruKitakaze/commit-msg-guardian/config"
	"github.com/AnruKitakaze/commit-msg-guardian/internal/parser"
	"github.com/AnruKitakaze/commit-msg-guardian/internal/rules"
)

var (
	headerType  = regexp.MustCompile(`^\w+$`)
	headerScope = regexp.MustCompile(`^[\w-]+(?:/[\w-]+)*$`)
	issueKey    = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9]*)-[1-9][0-9]*$`)
	projectKey  = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*$`)
)

type compiledChecks struct {
	typeRule, scopeRule, subjectRule, bodyRule rules.Checker
}

type compiledTrailer struct {
	name    string
	pattern *regexp.Regexp
}

type compiledFormat struct {
	config   config.Format
	header   *regexp.Regexp
	checks   compiledChecks
	trailers []compiledTrailer
}

// Validator validates one project policy, independently of how Git supplied a message.
type Validator struct {
	formats []compiledFormat
	base    config.CIBase
}

func New(cfg config.Config) (*Validator, error) {
	if cfg.Version == 0 {
		return nil, fmt.Errorf("config version is required (expected 1)")
	}
	if cfg.Version != 1 {
		return nil, fmt.Errorf("config version must be 1, got %d", cfg.Version)
	}
	if len(cfg.Commit.Formats) == 0 {
		return nil, fmt.Errorf("commit.formats must contain at least one format")
	}
	if cfg.CI.Commits.Base.Ref != "" && cfg.CI.Commits.Base.Env != "" {
		return nil, fmt.Errorf("ci.commits.base must set either ref or env, not both")
	}
	if cfg.CI.Commits.Base.Env != "" && !regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`).MatchString(cfg.CI.Commits.Base.Env) {
		return nil, fmt.Errorf("ci.commits.base.env must be an environment variable name")
	}
	v := &Validator{base: cfg.CI.Commits.Base}
	ids := make(map[string]bool)
	conditionalAuthors := make(map[string]bool)
	for _, f := range cfg.Commit.Formats {
		if f.ID == "" {
			return nil, fmt.Errorf("commit format id is required")
		}
		if ids[f.ID] {
			return nil, fmt.Errorf("commit format id %q is duplicated", f.ID)
		}
		ids[f.ID] = true
		if f.When != nil && f.When.AuthorEmail != "" {
			email := strings.ToLower(f.When.AuthorEmail)
			if conditionalAuthors[email] {
				return nil, fmt.Errorf("multiple conditional formats use author email %q", f.When.AuthorEmail)
			}
			conditionalAuthors[email] = true
		}
		compiled, err := compileFormat(f)
		if err != nil {
			return nil, fmt.Errorf("format %q: %w", f.ID, err)
		}
		v.formats = append(v.formats, compiled)
	}
	return v, nil
}

func (v *Validator) Base() config.CIBase { return v.base }

func (v *Validator) RequiresAuthor() bool {
	for _, f := range v.formats {
		if f.config.When != nil {
			return true
		}
	}
	return false
}

func compileFormat(f config.Format) (compiledFormat, error) {
	result := compiledFormat{config: f}
	typeCfg, scopeCfg := typePolicy(f), scopePolicy(f)
	if f.Type != nil && typeCfg.Allowed != nil && len(typeCfg.Allowed) == 0 {
		return result, fmt.Errorf("type.allowed must not be empty; omit it to allow every type")
	}
	if f.Scope != nil && scopeCfg.Allowed != nil && len(scopeCfg.Allowed) == 0 {
		return result, fmt.Errorf("scope.allowed must not be empty; omit it to allow every scope")
	}
	if f.Scope != nil {
		if scopeCfg.Brackets != "" && scopeCfg.Brackets != "none" && scopeCfg.Brackets != "round" && scopeCfg.Brackets != "square" {
			return result, fmt.Errorf("scope.brackets must be none, round, or square")
		}
		if f.Pattern == "" && f.Type != nil && (scopeCfg.Brackets == "" || scopeCfg.Brackets == "none") {
			return result, fmt.Errorf("scope.brackets must be round or square when type and scope are both configured; use pattern for another layout")
		}
		if f.Pattern == "" && f.Type == nil && !scopeRequired(scopeCfg) {
			return result, fmt.Errorf("scope.required: false needs a type block to make the header unambiguous")
		}
		if scopeCfg.Issue != nil {
			for _, project := range scopeCfg.Issue.Projects {
				if !projectKey.MatchString(project) {
					return result, fmt.Errorf("invalid scope.issue project %q", project)
				}
			}
		}
	}
	if f.Type == nil && len(f.Body.OptionalForTypes) != 0 {
		return result, fmt.Errorf("body.optional_for_types requires a type block")
	}
	if f.Pattern != "" {
		if len(f.Separators) != 0 || scopeCfg.Brackets != "" || scopeCfg.Issue != nil {
			return result, fmt.Errorf("pattern cannot be combined with separators, scope.brackets, or scope.issue")
		}
		if !strings.HasPrefix(f.Pattern, "^") || !strings.HasSuffix(f.Pattern, "$") {
			return result, fmt.Errorf("pattern must cover the full header (^...$)")
		}
		var err error
		result.header, err = regexp.Compile(f.Pattern)
		if err != nil {
			return result, fmt.Errorf("invalid pattern: %w", err)
		}
		if !slices.Contains(result.header.SubexpNames(), "subject") {
			return result, fmt.Errorf("pattern must contain a named subject group")
		}
		if f.Type != nil && !slices.Contains(result.header.SubexpNames(), "type") {
			return result, fmt.Errorf("pattern needs a named type group for type settings or body.optional_for_types")
		}
		if f.Scope != nil && !slices.Contains(result.header.SubexpNames(), "scope") {
			return result, fmt.Errorf("pattern needs a named scope group for scope settings")
		}
	} else {
		var err error
		result.header, err = buildHeader(f)
		if err != nil {
			return result, err
		}
	}
	if f.When != nil && f.When.AuthorEmail == "" {
		return result, fmt.Errorf("when.author_email must not be empty")
	}
	for _, value := range scopeCfg.Allowed {
		if value == "" {
			return result, fmt.Errorf("scope.allowed must not contain empty values")
		}
	}
	for _, value := range typeCfg.Allowed {
		if value == "" {
			return result, fmt.Errorf("type.allowed must not contain empty values")
		}
	}
	if err := validateLengths(f); err != nil {
		return result, err
	}
	var err error
	if result.checks.typeRule, err = rules.Compile(typeCfg.TextPolicy); err != nil {
		return result, fmt.Errorf("type: %w", err)
	}
	if result.checks.scopeRule, err = rules.Compile(scopeCfg.TextPolicy); err != nil {
		return result, fmt.Errorf("scope: %w", err)
	}
	if result.checks.subjectRule, err = rules.Compile(f.Subject.TextPolicy); err != nil {
		return result, fmt.Errorf("subject: %w", err)
	}
	if result.checks.bodyRule, err = rules.Compile(f.Body.TextPolicy); err != nil {
		return result, fmt.Errorf("body: %w", err)
	}
	if err := validateHeaderPolicies(f, result.checks); err != nil {
		return result, err
	}
	for _, trailer := range f.Trailers.Required {
		if !regexp.MustCompile(`^(?:[A-Za-z][A-Za-z0-9-]*|BREAKING CHANGE)$`).MatchString(trailer.Name) || trailer.ValuePattern == "" {
			return result, fmt.Errorf("required trailer needs a name and value_pattern")
		}
		pattern, err := regexp.Compile("^(?:" + trailer.ValuePattern + ")$")
		if err != nil {
			return result, fmt.Errorf("trailer %q: %w", trailer.Name, err)
		}
		result.trailers = append(result.trailers, compiledTrailer{trailer.Name, pattern})
	}
	return result, nil
}

func typePolicy(f config.Format) config.Type {
	if f.Type != nil {
		return *f.Type
	}
	return config.Type{}
}

func scopePolicy(f config.Format) config.Scope {
	if f.Scope != nil {
		return *f.Scope
	}
	return config.Scope{}
}

func scopeRequired(scope config.Scope) bool {
	return scope.Required == nil || *scope.Required
}

func buildHeader(f config.Format) (*regexp.Regexp, error) {
	if f.Type == nil && f.Scope == nil {
		if len(f.Separators) != 0 {
			return nil, fmt.Errorf("separators require a type or scope block")
		}
		return regexp.Compile(`^(?P<subject>.*)$`)
	}
	var prefix strings.Builder
	if f.Type != nil {
		prefix.WriteString(`(?P<type>\w+)`)
	}
	if f.Scope != nil {
		scope := `(?P<scope>[\w-]+(?:/[\w-]+)*)`
		if f.Scope.Issue != nil {
			scope = `(?P<scope>[A-Za-z][A-Za-z0-9]*-[1-9][0-9]*)`
		}
		switch f.Scope.Brackets {
		case "round":
			scope = `\(` + scope + `\)`
		case "square":
			scope = `\[` + scope + `\]`
		}
		if !scopeRequired(*f.Scope) {
			scope = `(?:` + scope + `)?`
		}
		prefix.WriteString(scope)
	}
	if len(f.Separators) > 0 {
		seen := make(map[string]bool)
		quoted := make([]string, 0, len(f.Separators))
		for _, separator := range f.Separators {
			if separator == "" || strings.TrimSpace(separator) != separator || strings.ContainsAny(separator, "\r\n") {
				return nil, fmt.Errorf("separators must contain nonempty values without surrounding whitespace or newlines")
			}
			if seen[separator] {
				return nil, fmt.Errorf("separators contains duplicate value %q", separator)
			}
			seen[separator] = true
			quoted = append(quoted, regexp.QuoteMeta(separator))
		}
		prefix.WriteString(`(?:` + strings.Join(quoted, `|`) + `)`)
	}
	return regexp.Compile(`^` + prefix.String() + `(?: (?P<subject>.*))?$`)
}

func validateHeaderPolicies(f config.Format, checks compiledChecks) error {
	typeCfg, scopeCfg := typePolicy(f), scopePolicy(f)
	for _, value := range typeCfg.Allowed {
		if f.Pattern == "" && !headerType.MatchString(value) {
			return fmt.Errorf("type.allowed value %q cannot match a type", value)
		}
		if err := checks.typeRule.Validate(value); err != nil {
			return fmt.Errorf("type.allowed value %q conflicts with type policy: %w", value, err)
		}
	}
	for _, value := range scopeCfg.Allowed {
		if f.Pattern == "" && scopeCfg.Issue == nil && !headerScope.MatchString(value) {
			return fmt.Errorf("scope.allowed value %q cannot match a scope", value)
		}
		if f.Pattern == "" && scopeCfg.Issue != nil && !issueKey.MatchString(value) {
			return fmt.Errorf("scope.allowed value %q cannot match an issue key", value)
		}
		if err := checks.scopeRule.Validate(value); err != nil {
			return fmt.Errorf("scope.allowed value %q conflicts with scope policy: %w", value, err)
		}
		if scopeCfg.Issue != nil && len(scopeCfg.Issue.Projects) > 0 && !slices.Contains(scopeCfg.Issue.Projects, issueKey.FindStringSubmatch(value)[1]) {
			return fmt.Errorf("scope.allowed value %q is not in scope.issue.projects", value)
		}
	}
	seenOptionalTypes := make(map[string]bool)
	for _, value := range f.Body.OptionalForTypes {
		if value == "" {
			return fmt.Errorf("body.optional_for_types must not contain empty values")
		}
		if seenOptionalTypes[value] {
			return fmt.Errorf("body.optional_for_types contains duplicate type %q", value)
		}
		seenOptionalTypes[value] = true
		if len(typeCfg.Allowed) > 0 && !slices.Contains(typeCfg.Allowed, value) {
			return fmt.Errorf("body.optional_for_types value %q is not in type.allowed", value)
		}
		if f.Pattern == "" && !headerType.MatchString(value) {
			return fmt.Errorf("body.optional_for_types value %q cannot match a type", value)
		}
	}
	if f.Pattern == "" {
		if f.Type != nil && len(typeCfg.Allowed) == 0 {
			if err := checks.typeRule.CheckAlphabet("a type", "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"); err != nil {
				return fmt.Errorf("type: %w", err)
			}
		}
		if f.Scope != nil && len(scopeCfg.Allowed) == 0 {
			alphabet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-/"
			if scopeCfg.Issue != nil {
				alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-"
			}
			if err := checks.scopeRule.CheckAlphabet("a scope", alphabet); err != nil {
				return fmt.Errorf("scope: %w", err)
			}
		}
	}
	return nil
}

func validateLengths(f config.Format) error {
	values := []struct {
		name  string
		value int
	}{
		{"subject.min_length", f.Subject.MinLength},
		{"subject.max_length", f.Subject.MaxLength},
		{"body.min_length", f.Body.MinLength},
		{"body.max_line_length", f.Body.MaxLineLength},
		{"body.max_total_length", f.Body.MaxTotalLength},
	}
	for _, item := range values {
		if item.value < 0 {
			return fmt.Errorf("%s must be non-negative", item.name)
		}
	}
	if f.Subject.MaxLength > 0 && f.Subject.MinLength > f.Subject.MaxLength {
		return fmt.Errorf("subject.min_length exceeds subject.max_length")
	}
	if f.Body.MaxTotalLength > 0 && f.Body.MinLength > f.Body.MaxTotalLength {
		return fmt.Errorf("body.min_length exceeds body.max_total_length")
	}
	return nil
}

func hasTextPolicy(policy config.TextPolicy) bool {
	return policy.Characters.Allowed != nil || policy.Characters.Required != nil || policy.Characters.Forbidden != nil ||
		policy.Structure != "" || policy.Case != "" || policy.TrailingPeriod != "" || policy.MaxLines != 0
}

type fields struct {
	typeName, scope, subject string
}

func (f compiledFormat) parse(header string) (fields, bool) {
	match := f.header.FindStringSubmatch(header)
	if match == nil {
		return fields{}, false
	}
	result := fields{}
	for i, name := range f.header.SubexpNames() {
		switch name {
		case "subject":
			result.subject = match[i]
		case "type":
			result.typeName = match[i]
		case "scope":
			result.scope = match[i]
		}
	}
	return result, true
}

// Validate accepts any matching general format, or the single author-specific
// format when its condition matches. A failed special format never falls back.
func (v *Validator) Validate(message, authorEmail string) error {
	parsed := parser.SplitMessage(message)
	var general []compiledFormat
	var selected []compiledFormat
	for _, f := range v.formats {
		if f.config.When == nil {
			general = append(general, f)
			continue
		}
		if authorEmail == "" {
			return fmt.Errorf("author email is required for conditional format %q", f.config.ID)
		}
		if strings.EqualFold(f.config.When.AuthorEmail, authorEmail) {
			selected = append(selected, f)
		}
	}
	if len(selected) > 1 {
		return fmt.Errorf("multiple conditional formats match author %q", authorEmail)
	}
	if len(selected) == 0 {
		selected = general
	}
	if len(selected) == 0 {
		return fmt.Errorf("no format applies to author %q", authorEmail)
	}
	var failures []string
	var expected []string
	for _, f := range selected {
		expected = append(expected, fmt.Sprintf("%s (%s)", f.config.ID, f.hint()))
		parts, ok := f.parse(parsed.Header)
		if !ok {
			continue
		}
		if err := f.validate(parts, parsed.Body, parsed.Trailers); err == nil {
			return nil
		} else {
			failures = append(failures, fmt.Sprintf("%s: %v", f.config.ID, err))
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("%s", strings.Join(failures, "; "))
	}
	return fmt.Errorf("header %q does not match expected format(s): %s", parsed.Header, strings.Join(expected, ", "))
}

func (f compiledFormat) hint() string {
	if f.config.Pattern != "" {
		return f.config.Pattern
	}
	var prefix strings.Builder
	if f.config.Type != nil {
		name := "type"
		if len(f.config.Type.Allowed) == 1 {
			name = f.config.Type.Allowed[0]
		}
		prefix.WriteString(name)
	}
	if f.config.Scope != nil {
		name := "scope"
		if f.config.Scope.Issue != nil {
			name = "TASK-1234"
			if len(f.config.Scope.Issue.Projects) > 0 {
				name = f.config.Scope.Issue.Projects[0] + "-1234"
			}
		} else if len(f.config.Scope.Allowed) == 1 {
			name = f.config.Scope.Allowed[0]
		}
		switch f.config.Scope.Brackets {
		case "round":
			name = "(" + name + ")"
		case "square":
			name = "[" + name + "]"
		}
		prefix.WriteString(name)
	}
	if len(f.config.Separators) > 0 {
		prefix.WriteString(f.config.Separators[0])
	}
	if prefix.Len() > 0 {
		prefix.WriteString(" ")
	}
	return prefix.String() + "subject"
}

func (f compiledFormat) validate(p fields, body, trailers string) error {
	cfg := f.config
	typeCfg, scopeCfg := typePolicy(cfg), scopePolicy(cfg)
	if p.subject == "" {
		return fmt.Errorf("subject must not be empty")
	}
	if cfg.Type != nil && p.typeName == "" {
		return fmt.Errorf("type is required")
	}
	if scopeCfg.Issue != nil && p.scope != "" && len(scopeCfg.Issue.Projects) > 0 {
		project := issueKey.FindStringSubmatch(p.scope)[1]
		if !slices.Contains(scopeCfg.Issue.Projects, project) {
			return fmt.Errorf("issue project %q is not allowed", project)
		}
	}
	if len(typeCfg.Allowed) > 0 && !slices.Contains(typeCfg.Allowed, p.typeName) {
		return fmt.Errorf("type %q is not allowed", p.typeName)
	}
	if cfg.Scope != nil && scopeRequired(scopeCfg) && p.scope == "" {
		return fmt.Errorf("scope is required")
	}
	if p.scope != "" && len(scopeCfg.Allowed) > 0 && !slices.Contains(scopeCfg.Allowed, p.scope) {
		return fmt.Errorf("scope %q is not allowed", p.scope)
	}
	if err := f.checks.typeRule.Validate(p.typeName); err != nil {
		return fmt.Errorf("type: %w", err)
	}
	if p.scope != "" {
		if err := f.checks.scopeRule.Validate(p.scope); err != nil {
			return fmt.Errorf("scope: %w", err)
		}
	}
	if err := f.checks.subjectRule.Validate(p.subject); err != nil {
		return fmt.Errorf("subject: %w", err)
	}
	if body != "" || !slices.Contains(cfg.Body.OptionalForTypes, p.typeName) {
		if err := f.checks.bodyRule.Validate(body); err != nil {
			return fmt.Errorf("body: %w", err)
		}
	}
	if err := validateSubject(cfg.Subject, p.subject); err != nil {
		return err
	}
	if err := validateBody(cfg.Body, p.typeName, body); err != nil {
		return err
	}
	footer := strings.Split(trailers, "\n")
	for _, trailer := range f.trailers {
		found := false
		for _, line := range footer {
			prefix := trailer.name + ": "
			if strings.HasPrefix(line, prefix) && trailer.pattern.MatchString(strings.TrimPrefix(line, prefix)) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("required trailer %q is missing or invalid", trailer.name)
		}
	}
	return nil
}

func validateSubject(cfg config.Text, value string) error {
	length := utf8.RuneCountInString(value)
	if length < cfg.MinLength || cfg.MaxLength > 0 && length > cfg.MaxLength {
		return fmt.Errorf("subject length %d is outside configured limits", length)
	}
	return nil
}

func validateBody(cfg config.Body, commitType, value string) error {
	optional := slices.Contains(cfg.OptionalForTypes, commitType)
	if cfg.Required && value == "" && !optional {
		return fmt.Errorf("body is required")
	}
	if value == "" && optional {
		return nil
	}
	length := utf8.RuneCountInString(value)
	if length < cfg.MinLength || cfg.MaxTotalLength > 0 && length > cfg.MaxTotalLength {
		return fmt.Errorf("body length %d is outside configured limits", length)
	}
	if cfg.MaxLineLength > 0 {
		for i, line := range strings.Split(value, "\n") {
			if utf8.RuneCountInString(line) > cfg.MaxLineLength {
				return fmt.Errorf("body line %d exceeds %d characters", i+1, cfg.MaxLineLength)
			}
		}
	}
	return nil
}
