package parser

import (
	"strings"
	"testing"
)

func TestParseCommitMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    *CommitMessage
		wantErr bool
	}{
		{
			name:    "valid commit without scope",
			message: "feat: add new feature",
			want: &CommitMessage{
				Type:        "feat",
				Description: "add new feature",
			},
			wantErr: false,
		},
		{
			name:    "valid commit with scope",
			message: "feat(scope): add new feature",
			want: &CommitMessage{
				Type:        "feat",
				Scope:       "scope",
				Description: "add new feature",
			},
			wantErr: false,
		},
		{
			name:    "valid breaking change without scope",
			message: "feat!: add new feature",
			want: &CommitMessage{
				Type:           "feat",
				BreakingChange: true,
				Description:    "add new feature",
			},
			wantErr: false,
		},
		{
			name:    "valid breaking change with scope",
			message: "feat(scope)!: add new feature",
			want: &CommitMessage{
				Type:           "feat",
				Scope:          "scope",
				BreakingChange: true,
				Description:    "add new feature",
			},
			wantErr: false,
		},
		{
			name:    "valid commit with slash-delimited scope",
			message: "feat(app/api): add endpoint",
			want: &CommitMessage{
				Type:        "feat",
				Scope:       "app/api",
				Description: "add endpoint",
			},
			wantErr: false,
		},
		{
			name: "valid commit with body",
			message: `feat(scope): add new feature

This is the body of the commit message
It can span multiple lines`,
			want: &CommitMessage{
				Type:        "feat",
				Scope:       "scope",
				Description: "add new feature",
				Body:        "This is the body of the commit message\nIt can span multiple lines",
			},
			wantErr: false,
		},
		{
			name: "valid commit with Git editor comments",
			message: `feat(scope): add new feature

This is the body of the commit message
# Please enter the commit message for your changes. Lines starting
# with '#' will be ignored, and an empty message aborts the commit.
#
# On branch main`,
			want: &CommitMessage{
				Type:        "feat",
				Scope:       "scope",
				Description: "add new feature",
				Body:        "This is the body of the commit message",
			},
			wantErr: false,
		},
		{
			name: "Git editor comments before header",
			message: `# Please enter the commit message for your changes.
feat: add new feature
# On branch main`,
			want: &CommitMessage{
				Type:        "feat",
				Description: "add new feature",
			},
			wantErr: false,
		},
		{
			name:    "invalid commit type",
			message: "invalid: not a valid type",
			wantErr: true,
		},
		{
			name:    "missing description",
			message: "feat(scope):",
			wantErr: true,
		},
		{
			name:    "invalid format",
			message: "just some text",
			wantErr: true,
		},
		{
			name:    "empty message",
			message: "",
			wantErr: true,
		},
		{
			name:    "invalid scope format",
			message: "feat[scope]: description",
			wantErr: true,
		},
		{
			name:    "invalid scope with empty slash segment",
			message: "feat(app//api): description",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCommitMessage(tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCommitMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Type != tt.want.Type {
					t.Errorf("ParseCommitMessage() Type = %v, want %v", got.Type, tt.want.Type)
				}
				if got.Scope != tt.want.Scope {
					t.Errorf("ParseCommitMessage() Scope = %v, want %v", got.Scope, tt.want.Scope)
				}
				if got.Description != tt.want.Description {
					t.Errorf("ParseCommitMessage() Description = %v, want %v", got.Description, tt.want.Description)
				}
				if got.BreakingChange != tt.want.BreakingChange {
					t.Errorf("ParseCommitMessage() BreakingChange = %v, want %v", got.BreakingChange, tt.want.BreakingChange)
				}
				if !strings.EqualFold(got.Body, tt.want.Body) {
					t.Errorf("ParseCommitMessage() Body = %v, want %v", got.Body, tt.want.Body)
				}
			}
		})
	}
}

func TestRemoveCommentLines(t *testing.T) {
	message := "feat: add new feature\n # This is content, not a Git comment\n# This is a Git comment\n"
	want := "feat: add new feature\n # This is content, not a Git comment\n"

	if got := removeCommentLines(message); got != want {
		t.Errorf("removeCommentLines() = %q, want %q", got, want)
	}
}

func TestValidateWithRules(t *testing.T) {
	tests := []struct {
		name       string
		message    *CommitMessage
		typeRules  []string
		scopeRules []string
		descRules  []string
		bodyRules  []string
		wantErr    bool
	}{
		{
			name: "all valid with default rules",
			message: &CommitMessage{
				Type:        "feat",
				Scope:       "T-1",
				Description: "add new feature",
			},
			typeRules:  []string{"allowLatin"},
			scopeRules: []string{"allowScope"},
			descRules:  []string{"noCyrillic"},
			bodyRules:  nil,
			wantErr:    false,
		},
		{
			name: "slash-delimited scope",
			message: &CommitMessage{
				Type:        "feat",
				Scope:       "app/api",
				Description: "add new feature",
			},
			typeRules:  []string{"allowLatin"},
			scopeRules: []string{"allowPathScope"},
			descRules:  []string{"noCyrillic"},
			bodyRules:  nil,
			wantErr:    false,
		},
		{
			name: "capitalized description",
			message: &CommitMessage{
				Type:        "feat",
				Scope:       "T-1",
				Description: "Add new feature",
			},
			typeRules:  []string{"allowLatin"},
			scopeRules: []string{"allowScope"},
			descRules:  []string{"capitalized"},
			bodyRules:  nil,
			wantErr:    false,
		},
		{
			name: "non-capitalized description",
			message: &CommitMessage{
				Type:        "feat",
				Scope:       "T-1",
				Description: "add new feature",
			},
			typeRules:  []string{"allowLatin"},
			scopeRules: []string{"allowScope"},
			descRules:  []string{"capitalized"},
			bodyRules:  nil,
			wantErr:    true,
		},
		{
			name: "description with required trailing period",
			message: &CommitMessage{
				Type:        "feat",
				Scope:       "T-1",
				Description: "Add new feature.",
			},
			typeRules:  []string{"allowLatin"},
			scopeRules: []string{"allowScope"},
			descRules:  []string{"trailingPeriod"},
			bodyRules:  nil,
			wantErr:    false,
		},
		{
			name: "description without required trailing period",
			message: &CommitMessage{
				Type:        "feat",
				Scope:       "T-1",
				Description: "Add new feature",
			},
			typeRules:  []string{"allowLatin"},
			scopeRules: []string{"allowScope"},
			descRules:  []string{"trailingPeriod"},
			bodyRules:  nil,
			wantErr:    true,
		},
		{
			name: "body without trailing period",
			message: &CommitMessage{
				Type:        "feat",
				Scope:       "T-1",
				Description: "add new feature",
				Body:        "This body stays on one line",
			},
			typeRules:  []string{"allowLatin"},
			scopeRules: []string{"allowScope"},
			descRules:  []string{"noCyrillic"},
			bodyRules:  []string{"noTrailingPeriod"},
			wantErr:    false,
		},
		{
			name: "body with forbidden trailing period",
			message: &CommitMessage{
				Type:        "feat",
				Scope:       "T-1",
				Description: "add new feature",
				Body:        "This body stays on one line.",
			},
			typeRules:  []string{"allowLatin"},
			scopeRules: []string{"allowScope"},
			descRules:  []string{"noCyrillic"},
			bodyRules:  []string{"noTrailingPeriod"},
			wantErr:    true,
		},
		{
			name: "one-line body",
			message: &CommitMessage{
				Type:        "feat",
				Scope:       "T-1",
				Description: "add new feature",
				Body:        "This body stays on one line",
			},
			typeRules:  []string{"allowLatin"},
			scopeRules: []string{"allowScope"},
			descRules:  []string{"noCyrillic"},
			bodyRules:  []string{"oneLine"},
			wantErr:    false,
		},
		{
			name: "multi-line body",
			message: &CommitMessage{
				Type:        "feat",
				Scope:       "T-1",
				Description: "add new feature",
				Body:        "First body line\nSecond body line",
			},
			typeRules:  []string{"allowLatin"},
			scopeRules: []string{"allowScope"},
			descRules:  []string{"noCyrillic"},
			bodyRules:  []string{"oneLine"},
			wantErr:    true,
		},
		{
			name: "cyrillic in description",
			message: &CommitMessage{
				Type:        "feat",
				Scope:       "T-1",
				Description: "добавить фичу",
			},
			typeRules:  []string{"allowLatin"},
			scopeRules: []string{"allowScope"},
			descRules:  []string{"noCyrillic"},
			bodyRules:  nil,
			wantErr:    true,
		},
		{
			name: "invalid scope format",
			message: &CommitMessage{
				Type:        "feat",
				Scope:       "-T1",
				Description: "add new feature",
			},
			typeRules:  []string{"allowLatin"},
			scopeRules: []string{"allowScope"},
			descRules:  []string{"noCyrillic"},
			bodyRules:  nil,
			wantErr:    true,
		},
		{
			name: "multiple rules for type",
			message: &CommitMessage{
				Type:        "feat123",
				Scope:       "T-1",
				Description: "add new feature",
			},
			typeRules:  []string{"allowLatin", "noDigits"},
			scopeRules: []string{"allowScope"},
			descRules:  []string{"noCyrillic"},
			bodyRules:  nil,
			wantErr:    true,
		},
		{
			name: "invalid rule name",
			message: &CommitMessage{
				Type:        "feat",
				Scope:       "T-1",
				Description: "add new feature",
			},
			typeRules:  []string{"nonexistentRule"},
			scopeRules: []string{"allowScope"},
			descRules:  []string{"noCyrillic"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.message.ValidateWithRules(tt.typeRules, tt.scopeRules, tt.descRules, tt.bodyRules)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWithRules() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateLengthLimits(t *testing.T) {
	tests := []struct {
		name             string
		message          *CommitMessage
		descriptionLimit int
		bodyLimit        int
		wantErr          bool
		wantErrText      string
	}{
		{
			name: "disabled limits",
			message: &CommitMessage{
				Description: "Any description length",
				Body:        "Any body length",
			},
			descriptionLimit: 0,
			bodyLimit:        0,
			wantErr:          false,
		},
		{
			name: "description under limit",
			message: &CommitMessage{
				Description: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			descriptionLimit: 60,
			wantErr:          false,
		},
		{
			name: "description at limit",
			message: &CommitMessage{
				Description: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			descriptionLimit: 60,
			wantErr:          false,
		},
		{
			name: "description over limit",
			message: &CommitMessage{
				Description: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			descriptionLimit: 60,
			wantErr:          true,
			wantErrText:      "description must be no longer than 60 characters, got 61",
		},
		{
			name: "body under limit",
			message: &CommitMessage{
				Body: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			bodyLimit: 72,
			wantErr:   false,
		},
		{
			name: "body at limit",
			message: &CommitMessage{
				Body: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			bodyLimit: 72,
			wantErr:   false,
		},
		{
			name: "body over limit",
			message: &CommitMessage{
				Body: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			bodyLimit:   72,
			wantErr:     true,
			wantErrText: "body must be no longer than 72 characters, got 73",
		},
		{
			name: "unicode rune length",
			message: &CommitMessage{
				Description: "ЖЖЖ",
				Body:        "ЖЖЖЖ",
			},
			descriptionLimit: 4,
			bodyLimit:        4,
			wantErr:          false,
		},
		{
			name: "unicode rune length over limit",
			message: &CommitMessage{
				Description: "ЖЖЖЖЖ",
			},
			descriptionLimit: 4,
			wantErr:          true,
			wantErrText:      "description must be no longer than 4 characters, got 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.message.ValidateLengthLimits(tt.descriptionLimit, tt.bodyLimit)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLengthLimits() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErrText != "" && err.Error() != tt.wantErrText {
				t.Errorf("ValidateLengthLimits() error = %v, want %v", err, tt.wantErrText)
			}
		})
	}
}
