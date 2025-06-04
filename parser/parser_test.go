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
				if !strings.EqualFold(got.Body, tt.want.Body) {
					t.Errorf("ParseCommitMessage() Body = %v, want %v", got.Body, tt.want.Body)
				}
			}
		})
	}
}

func TestValidateWithRules(t *testing.T) {
	tests := []struct {
		name       string
		message    *CommitMessage
		typeRules  []string
		scopeRules []string
		descRules  []string
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
			wantErr:    false,
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
			err := tt.message.ValidateWithRules(tt.typeRules, tt.scopeRules, tt.descRules)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWithRules() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
