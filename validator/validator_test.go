package validator

import (
	"strings"
	"testing"

	"github.com/AnruKitakaze/commit-msg-guardian/config"
)

func policy(formats ...config.Format) config.Config {
	return config.Config{Version: 1, Commit: config.CommitConfig{Formats: formats}}
}

func optionalScope() *bool { value := false; return &value }

func TestComposedFormatsAndRenovatePriority(t *testing.T) {
	v, err := New(policy(
		config.Format{ID: "jira", Scope: &config.Scope{Brackets: "square", Issue: &config.Issue{Projects: []string{"TASK"}}}},
		config.Format{ID: "chore", Type: &config.Type{Allowed: []string{"chore"}}, Scope: &config.Scope{Brackets: "round", Required: optionalScope()}, Separators: []string{":"}},
		config.Format{ID: "renovate", Type: &config.Type{Allowed: []string{"chore", "fix"}}, Scope: &config.Scope{Brackets: "round", Allowed: []string{"deps"}}, Separators: []string{":"}, When: &config.When{AuthorEmail: "bot@example.com"}},
	))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		message, author string
		valid           bool
	}{
		{"[TASK-1234] do something\n\noptional text", "human@example.com", true},
		{"chore: bump packageX", "human@example.com", true},
		{"chore(deps): update packageX", "human@example.com", true},
		{"fix(deps): update module google.golang.org/grpc to v1.23.4 [security]", "bot@example.com", true},
		{"chore(deps): update module golang.org/x/text to v0.12.0 [security]", "bot@example.com", true},
		{"chore: bump packageX", "bot@example.com", false},
		{"fix(deps): update packageX", "human@example.com", false},
		{"[OTHER-1234] do something", "human@example.com", false},
		{"(TASK-1234) do something", "human@example.com", false},
		{"[TASK-1234]", "human@example.com", false},
	} {
		if err := v.Validate(tc.message, tc.author); (err == nil) != tc.valid {
			t.Errorf("Validate(%q, %q) = %v, want valid %v", tc.message, tc.author, err, tc.valid)
		}
	}
	if err := v.Validate("[TASK-1234]", "human@example.com"); err == nil || !strings.Contains(err.Error(), "subject must not be empty") {
		t.Errorf("empty subject diagnostic = %v", err)
	}
}

func TestIssueScopeStylesAndMixedType(t *testing.T) {
	for _, tc := range []struct {
		name      string
		format    config.Format
		good, bad string
	}{
		{"square", config.Format{ID: "square", Scope: &config.Scope{Brackets: "square", Issue: &config.Issue{Projects: []string{"TASK"}}}}, "[TASK-1234] subject", "(TASK-1234) subject"},
		{"round", config.Format{ID: "round", Scope: &config.Scope{Brackets: "round", Issue: &config.Issue{Projects: []string{"TASK"}}}}, "(TASK-1234) subject", "[TASK-1234] subject"},
		{"colon", config.Format{ID: "colon", Scope: &config.Scope{Issue: &config.Issue{Projects: []string{"TASK"}}}, Separators: []string{":"}}, "TASK-1234: subject", "[TASK-1234] subject"},
		{"type and task", config.Format{ID: "typed", Type: &config.Type{Allowed: []string{"feat"}}, Scope: &config.Scope{Brackets: "round", Issue: &config.Issue{Projects: []string{"TASK"}}}, Separators: []string{":"}}, "feat(TASK-1234): subject", "feat(OTHER-1234): subject"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v, err := New(policy(tc.format))
			if err != nil {
				t.Fatal(err)
			}
			if err := v.Validate(tc.good, ""); err != nil {
				t.Errorf("valid = %v", err)
			}
			if err := v.Validate(tc.bad, ""); err == nil {
				t.Errorf("%q should fail", tc.bad)
			}
		})
	}
}

func TestBothBracketStylesNeedTwoFormats(t *testing.T) {
	v, err := New(policy(
		config.Format{ID: "square", Scope: &config.Scope{Brackets: "square", Issue: &config.Issue{Projects: []string{"TASK"}}}},
		config.Format{ID: "round", Scope: &config.Scope{Brackets: "round", Issue: &config.Issue{Projects: []string{"TASK"}}}},
	))
	if err != nil {
		t.Fatal(err)
	}
	for _, message := range []string{"[TASK-1234] subject", "(TASK-1234) subject"} {
		if err := v.Validate(message, ""); err != nil {
			t.Errorf("Validate(%q) = %v", message, err)
		}
	}
}

func TestSeparatorChoices(t *testing.T) {
	v, err := New(policy(config.Format{ID: "conventional", Type: &config.Type{Allowed: []string{"feat", "fix"}}, Scope: &config.Scope{Brackets: "round", Required: optionalScope()}, Separators: []string{":", "!:"}}))
	if err != nil {
		t.Fatal(err)
	}
	for _, message := range []string{"feat: add feature", "feat!: replace API", "fix(api): correct output", "feat(api)!: replace API", "feat: replace API\n\nBREAKING CHANGE: migrate clients"} {
		if err := v.Validate(message, ""); err != nil {
			t.Errorf("Validate(%q) = %v", message, err)
		}
	}
	for _, message := range []string{"feat! : replace API", "feat(!api): replace API", "feat(api):", "feat(api)!: "} {
		if err := v.Validate(message, ""); err == nil {
			t.Errorf("Validate(%q) should fail", message)
		}
	}
}

func TestTextPoliciesBodyTrailerAndPattern(t *testing.T) {
	format := config.Format{
		ID: "legacy", Type: &config.Type{Allowed: []string{"feat", "docs"}}, Scope: &config.Scope{Brackets: "round", Required: optionalScope(), TextPolicy: config.TextPolicy{Structure: "slash-separated"}}, Separators: []string{":"},
		Subject:  config.Text{MaxLength: 20, TextPolicy: config.TextPolicy{Characters: config.Characters{Forbidden: []string{"Cyrillic"}}, Case: "upper-first-letter", TrailingPeriod: "require"}},
		Body:     config.Body{Required: true, OptionalForTypes: []string{"docs"}, MaxTotalLength: 50, TextPolicy: config.TextPolicy{TrailingPeriod: "forbid"}},
		Trailers: config.Trailers{Required: []config.RequiredTrailer{{Name: "Refs", ValuePattern: `TASK-[1-9][0-9]*`}}},
	}
	v, err := New(policy(format))
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Validate("feat(app/api): Add.\n\nExplanation\n\nRefs: TASK-1234", ""); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ message, want string }{
		{"feat(app/api): add.\n\nExplanation\n\nRefs: TASK-1234", "upper"},
		{"feat(app//api): Add.\n\nExplanation\n\nRefs: TASK-1234", "expected format"},
		{"feat(app/api): Add.\n\n# real body", "body is required"},
		{"docs: Update.", "Refs"},
		{"feat(app/api): Add.\n\nExplanation\n\nRefs: OTHER-1234", "Refs"},
	} {
		err := v.Validate(tc.message, "")
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("Validate(%q) = %v, want %q", tc.message, err, tc.want)
		}
	}
	custom := config.Format{ID: "custom", Pattern: `^TASK-(?P<issue>[1-9][0-9]*) :: (?P<subject>.+)$`, Subject: config.Text{MaxLength: 10}}
	v, err = New(policy(custom))
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Validate("TASK-42 :: short", ""); err != nil {
		t.Fatal(err)
	}
	if err := v.Validate("TASK-42 :: a long subject", ""); err == nil {
		t.Fatal("subject length should apply to named group")
	}
	v, err = New(policy(config.Format{ID: "custom-parts", Pattern: `^(?P<type>[a-z]+)\[(?P<scope>[a-z]+)\]: (?P<subject>.+)$`, Type: &config.Type{Allowed: []string{"feat"}}, Scope: &config.Scope{Allowed: []string{"api"}}}))
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Validate("feat[api]: add endpoint", ""); err != nil {
		t.Fatal(err)
	}
	if err := v.Validate("fix[api]: add endpoint", ""); err == nil || !strings.Contains(err.Error(), "type") {
		t.Errorf("custom pattern must retain type checks: %v", err)
	}
	v, err = New(policy(config.Format{ID: "optional-capture", Pattern: `^(?:(?P<type>feat): )?(?P<subject>.+)$`, Type: &config.Type{}}))
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Validate("add feature", ""); err == nil || !strings.Contains(err.Error(), "type is required") {
		t.Errorf("configured type must not disappear through an optional pattern group: %v", err)
	}
	v, err = New(policy(config.Format{ID: "plain", Subject: config.Text{MaxLength: 13}}))
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Validate("Plain subject", ""); err != nil {
		t.Fatal(err)
	}
	if err := v.Validate("", ""); err == nil {
		t.Fatal("empty subject should fail")
	}
}

func TestUnicodeLengths(t *testing.T) {
	v, err := New(policy(config.Format{ID: "unicode", Type: &config.Type{Allowed: []string{"chore"}}, Separators: []string{":"}, Subject: config.Text{MaxLength: 2}, Body: config.Body{MaxTotalLength: 3}}))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ message, want string }{
		{"chore: ЖЖ\n\nЖЖЖ", ""},
		{"chore: ЖЖЖ\n\nЖЖЖ", "subject length"},
		{"chore: ЖЖ\n\nЖЖЖЖ", "body length"},
	} {
		err := v.Validate(tc.message, "")
		if tc.want == "" && err != nil || tc.want != "" && (err == nil || !strings.Contains(err.Error(), tc.want)) {
			t.Errorf("Validate(%q) = %v, want %q", tc.message, err, tc.want)
		}
	}
}

func TestBodyRulesDoNotCountTrailers(t *testing.T) {
	v, err := New(policy(config.Format{
		ID: "with-trailer", Type: &config.Type{Allowed: []string{"feat"}}, Separators: []string{":"},
		Body:     config.Body{Required: true, MaxTotalLength: 5},
		Trailers: config.Trailers{Required: []config.RequiredTrailer{{Name: "BREAKING CHANGE", ValuePattern: ".+"}}},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Validate("feat: change\n\nshort\n\nBREAKING CHANGE: migrate to v2", ""); err != nil {
		t.Fatalf("trailer should not count toward body length: %v", err)
	}
	if err := v.Validate("feat: change\n\nBREAKING CHANGE: migrate to v2", ""); err == nil || !strings.Contains(err.Error(), "body is required") {
		t.Fatalf("trailer-only message should not satisfy body.required: %v", err)
	}
}

func TestInvalidConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name string
		cfg  config.Config
		want string
	}{
		{"missing version", config.Config{}, "version is required"},
		{"missing formats", config.Config{Version: 1}, "commit.formats must contain"},
		{"missing id", policy(config.Format{}), "format id is required"},
		{"bad brackets", policy(config.Format{ID: "x", Scope: &config.Scope{Brackets: "curly"}}), "scope.brackets"},
		{"bare scope after type", policy(config.Format{ID: "x", Type: &config.Type{}, Scope: &config.Scope{}}), "must be round or square"},
		{"optional scope without type", policy(config.Format{ID: "x", Scope: &config.Scope{Required: optionalScope()}}), "needs a type block"},
		{"pattern missing type group", policy(config.Format{ID: "x", Pattern: `^(?P<subject>.+)$`, Type: &config.Type{Allowed: []string{"feat"}}}), "named type group"},
		{"pattern missing scope group", policy(config.Format{ID: "x", Pattern: `^(?P<subject>.+)$`, Scope: &config.Scope{}}), "named scope group"},
		{"pattern and separator", policy(config.Format{ID: "x", Pattern: `^(?P<subject>.+)$`, Separators: []string{":"}}), "pattern cannot be combined"},
		{"empty type allowlist", policy(config.Format{ID: "x", Type: &config.Type{Allowed: []string{}}}), "type.allowed must not be empty"},
		{"empty scope allowlist", policy(config.Format{ID: "x", Scope: &config.Scope{Allowed: []string{}}}), "scope.allowed must not be empty"},
		{"Cyrillic required in type", policy(config.Format{ID: "x", Type: &config.Type{TextPolicy: config.TextPolicy{Characters: config.Characters{Required: []string{"Cyrillic"}}}}}), "cannot occur in a type"},
		{"Cyrillic required in scope", policy(config.Format{ID: "x", Scope: &config.Scope{TextPolicy: config.TextPolicy{Characters: config.Characters{Required: []string{"Cyrillic"}}}}}), "cannot occur in a scope"},
		{"type allowlist conflicts", policy(config.Format{ID: "x", Type: &config.Type{Allowed: []string{"feat"}, TextPolicy: config.TextPolicy{Characters: config.Characters{Required: []string{"digit"}}}}}), "type.allowed value"},
		{"scope allowlist conflicts", policy(config.Format{ID: "x", Scope: &config.Scope{Allowed: []string{"api_core"}, TextPolicy: config.TextPolicy{Structure: "flat"}}}), "scope.allowed value"},
		{"scope cannot match header", policy(config.Format{ID: "x", Scope: &config.Scope{Allowed: []string{"api?core"}}}), "cannot match a scope"},
		{"exception outside types", policy(config.Format{ID: "x", Type: &config.Type{Allowed: []string{"feat"}}, Body: config.Body{OptionalForTypes: []string{"docs"}}}), "not in type.allowed"},
		{"exception without type", policy(config.Format{ID: "x", Body: config.Body{OptionalForTypes: []string{"docs"}}}), "requires a type block"},
		{"duplicate separator", policy(config.Format{ID: "x", Type: &config.Type{}, Separators: []string{":", ":"}}), "duplicate"},
		{"separator without prefix", policy(config.Format{ID: "x", Separators: []string{":"}}), "require a type or scope"},
		{"required period conflicts", policy(config.Format{ID: "x", Subject: config.Text{TextPolicy: config.TextPolicy{Characters: config.Characters{Allowed: []string{"Latin"}}, TrailingPeriod: "require"}}}), "period is not permitted"},
		{"negative length", policy(config.Format{ID: "x", Body: config.Body{MaxTotalLength: -1}}), "non-negative"},
		{"unknown character group", policy(config.Format{ID: "x", Body: config.Body{TextPolicy: config.TextPolicy{Characters: config.Characters{Allowed: []string{"unknown"}}}}}), "unknown character group"},
		{"ambiguous base", config.Config{Version: 1, Commit: config.CommitConfig{Formats: []config.Format{{ID: "x"}}}, CI: config.CIConfig{Commits: config.CICommits{Base: config.CIBase{Ref: "main", Env: "BASE"}}}}, "either ref or env"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(tc.cfg)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("New() = %v, want %q", err, tc.want)
			}
		})
	}
}
