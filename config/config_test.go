package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadRejectsUnknownAndMultipleDocuments(t *testing.T) {
	for _, tc := range []struct {
		name, yaml, want string
	}{
		{"empty config", "", "version and commit.formats are required"},
		{"unknown nested key", "version: 1\ncommit:\n  formats:\n    - id: jira\n      typo: true\n", "typo"},
		{"old kind key", "version: 1\ncommit:\n  formats:\n    - id: x\n      kind: conventional\n", "kind"},
		{"old checks key", "version: 1\ncommit:\n  formats:\n    - id: x\n      checks:\n        subject: [noCyrillic]\n", "checks"},
		{"old types key", "version: 1\ncommit:\n  formats:\n    - id: x\n      types: [feat]\n", "types"},
		{"old syntax key", "version: 1\ncommit:\n  formats:\n    - id: x\n      scope:\n        syntax: scope-segment\n", "syntax"},
		{"duplicate key", "version: 1\nversion: 1\n", "already defined"},
		{"multiple documents", "version: 1\n---\nversion: 1\n", "exactly one"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(path, []byte(tc.yaml), 0600); err != nil {
				t.Fatal(err)
			}
			_, err := Load(path)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Load() error = %v, want containing %q", err, tc.want)
			}
		})
	}
}

func TestLoadCharacterPolicy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	contents := "version: 1\ncommit:\n  formats:\n    - id: x\n      type:\n        allowed: [feat]\n      scope:\n        brackets: round\n        required: false\n        structure: slash-separated\n      separators: [':', '!:']\n      subject:\n        characters:\n          allowed: [Latin, digit, whitespace]\n          required: [Latin]\n"
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	format := cfg.Commit.Formats[0]
	if format.Type == nil || len(format.Type.Allowed) != 1 || format.Type.Allowed[0] != "feat" || format.Scope == nil || format.Scope.Structure != "slash-separated" || format.Scope.Required == nil || *format.Scope.Required || len(format.Separators) != 2 ||
		len(format.Subject.Characters.Required) != 1 || format.Subject.Characters.Required[0] != "Latin" {
		t.Fatalf("decoded format = %+v", format)
	}
}
