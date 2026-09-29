package rules

import (
	"strings"
	"testing"

	"github.com/AnruKitakaze/commit-msg-guardian/config"
)

func TestCharacterGroups(t *testing.T) {
	for _, tc := range []struct {
		name       string
		characters config.Characters
		accepted   []string
		rejected   []string
	}{
		{
			name: "Latin text with digits",
			characters: config.Characters{
				Allowed: []string{"Latin", "digit", "whitespace", "punctuation"}, Required: []string{"Latin"},
			},
			accepted: []string{"Update API v2", "Fix café"},
			rejected: []string{"Обновить API", "123", "Add C++ support", "Fix e\u0301mail"},
		},
		{
			name: "symbols and combining marks explicitly allowed",
			characters: config.Characters{
				Allowed: []string{"Latin", "whitespace", "punctuation", "symbol", "mark"}, Required: []string{"Latin"},
			},
			accepted: []string{"Add C++ support", "Fix e\u0301mail"},
			rejected: []string{"Обновить C++"},
		},
		{
			name: "Russian with English terms",
			characters: config.Characters{
				Allowed: []string{"Cyrillic", "Latin", "digit", "whitespace", "punctuation"}, Required: []string{"Cyrillic"},
			},
			accepted: []string{"Исправить API flow v2"},
			rejected: []string{"Fix API flow", "改善 API flow"},
		},
		{
			name: "Han characters with English terms",
			characters: config.Characters{
				Allowed: []string{"Han", "Latin", "digit", "whitespace", "punctuation"}, Required: []string{"Han"},
			},
			accepted: []string{"更新 API v2"},
			rejected: []string{"Update API v2", "Обновить API", "更新を確認"},
		},
		{
			name: "every required group must occur",
			characters: config.Characters{
				Allowed: []string{"Cyrillic", "Latin", "whitespace"}, Required: []string{"Cyrillic", "Latin"},
			},
			accepted: []string{"Исправить API"},
			rejected: []string{"Исправить", "Fix API"},
		},
		{
			name:       "forbid Cyrillic only",
			characters: config.Characters{Forbidden: []string{"Cyrillic"}},
			accepted:   []string{"Fix API", "改善 API", "Fix \u0483 mark"},
			rejected:   []string{"Исправить API"},
		},
		{
			name:       "forbid Han only",
			characters: config.Characters{Forbidden: []string{"Han"}},
			accepted:   []string{"Fix API", "Исправить API"},
			rejected:   []string{"更新 API"},
		},
		{
			name:       "forbid combining marks explicitly",
			characters: config.Characters{Forbidden: []string{"mark"}},
			accepted:   []string{"Fix café"},
			rejected:   []string{"Fix e\u0301mail", "Fix \u0483 mark"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			checker, err := Compile(config.TextPolicy{Characters: tc.characters})
			if err != nil {
				t.Fatal(err)
			}
			for _, value := range tc.accepted {
				if err := checker.Validate(value); err != nil {
					t.Errorf("accepted %q: %v", value, err)
				}
			}
			for _, value := range tc.rejected {
				if err := checker.Validate(value); err == nil {
					t.Errorf("rejected %q was accepted", value)
				}
			}
		})
	}
}

func TestLegacyCharacterMappings(t *testing.T) {
	for _, tc := range []struct {
		name           string
		characters     config.Characters
		valid, invalid string
	}{
		{"noCyrillic", config.Characters{Forbidden: []string{"Cyrillic"}}, "Hello", "Привет"},
		{"noLatin", config.Characters{Forbidden: []string{"Latin"}}, "Привет", "Hello"},
		{"noDigits", config.Characters{Forbidden: []string{"digit"}}, "Hello", "Hello 123"},
		{"cyrillicOnly", config.Characters{Allowed: []string{"Cyrillic", "whitespace", "punctuation"}}, "Привет, мир!", "Привет 123"},
		{"latinOnly", config.Characters{Allowed: []string{"Latin", "whitespace", "punctuation"}}, "Hello, world!", "Hello 123"},
		{"digitsOnly", config.Characters{Allowed: []string{"digit", "whitespace", "punctuation"}}, "123, 456!", "Hello 123"},
		{"allowLatin and allowDigits", config.Characters{Allowed: []string{"Latin", "digit", "whitespace", "punctuation"}}, "Hello 123!", "Привет 123"},
		{"allowCyrillic", config.Characters{Allowed: []string{"Cyrillic", "digit", "whitespace", "punctuation"}}, "Привет 123!", "Hello 123"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			checker, err := Compile(config.TextPolicy{Characters: tc.characters})
			if err != nil {
				t.Fatal(err)
			}
			if err := checker.Validate(tc.valid); err != nil {
				t.Fatalf("valid example %q: %v", tc.valid, err)
			}
			if err := checker.Validate(tc.invalid); err == nil {
				t.Fatalf("invalid example %q was accepted", tc.invalid)
			}
		})
	}
}

func TestStructuredTextChecks(t *testing.T) {
	checker, err := Compile(config.TextPolicy{
		Structure: "slash-separated", Case: "upper-first-letter", TrailingPeriod: "forbid", MaxLines: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := checker.Validate("App/api"); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"App//api", "app/api", "App/api.", "App/api\nNext"} {
		if err := checker.Validate(value); err == nil {
			t.Errorf("Validate(%q) unexpectedly passed", value)
		}
	}
}

func TestScopeStructures(t *testing.T) {
	for _, tc := range []struct {
		structure string
		accepted  []string
		rejected  []string
	}{
		{"flat", []string{"api", "api-v2"}, []string{"api/client", "api_core", "-api", "api-"}},
		{"slash-separated", []string{"api", "api/client-v2"}, []string{"api//client", "api_core", "/api", "api/"}},
	} {
		t.Run(tc.structure, func(t *testing.T) {
			checker, err := Compile(config.TextPolicy{Structure: tc.structure})
			if err != nil {
				t.Fatal(err)
			}
			for _, value := range tc.accepted {
				if err := checker.Validate(value); err != nil {
					t.Errorf("expected %q to pass: %v", value, err)
				}
			}
			for _, value := range tc.rejected {
				if err := checker.Validate(value); err == nil {
					t.Errorf("expected %q to fail", value)
				}
			}
		})
	}
}

func TestInvalidCharacterPolicy(t *testing.T) {
	for _, tc := range []struct {
		name   string
		policy config.TextPolicy
		want   string
	}{
		{"unknown group", config.TextPolicy{Characters: config.Characters{Allowed: []string{"not-a-group"}}}, "unknown character group"},
		{"empty allowlist", config.TextPolicy{Characters: config.Characters{Allowed: []string{}}}, "must not be empty"},
		{"required not allowed", config.TextPolicy{Characters: config.Characters{Allowed: []string{"Latin"}, Required: []string{"Cyrillic"}}}, "cannot occur under characters.allowed"},
		{"allowed and forbidden", config.TextPolicy{Characters: config.Characters{Allowed: []string{"Latin"}, Forbidden: []string{"Latin"}}}, "completely forbidden"},
		{"script forbidden by broad group", config.TextPolicy{Characters: config.Characters{Allowed: []string{"Latin"}, Forbidden: []string{"letter"}}}, "completely forbidden"},
		{"required script forbidden by broad group", config.TextPolicy{Characters: config.Characters{Required: []string{"Latin"}, Forbidden: []string{"letter"}}}, "forbidden"},
		{"required period not allowed", config.TextPolicy{Characters: config.Characters{Allowed: []string{"Latin"}}, TrailingPeriod: "require"}, "period is not permitted"},
		{"required period forbidden", config.TextPolicy{Characters: config.Characters{Forbidden: []string{"punctuation"}}, TrailingPeriod: "require"}, "period is not permitted"},
		{"required period conflicts with flat structure", config.TextPolicy{Structure: "flat", TrailingPeriod: "require"}, "conflicts with structure"},
		{"flat structure needs alphanumeric", config.TextPolicy{Structure: "flat", Characters: config.Characters{Allowed: []string{"punctuation"}}}, "needs a Latin letter or digit"},
		{"slash-separated structure cannot require whitespace", config.TextPolicy{Structure: "slash-separated", Characters: config.Characters{Required: []string{"whitespace"}}}, "cannot occur in slash-separated structure"},
		{"invalid structure", config.TextPolicy{Structure: "path"}, "structure must be"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Compile(tc.policy)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Compile() = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestFixedHeaderAlphabet(t *testing.T) {
	for _, tc := range []struct {
		name, alphabet, want string
		policy               config.TextPolicy
	}{
		{"Cyrillic cannot appear in ASCII type", "abcABC123_", "cyrillic", config.TextPolicy{Characters: config.Characters{Required: []string{"Cyrillic"}}}},
		{"whitespace-only policy excludes ASCII type", "abcABC123_", "permits no characters", config.TextPolicy{Characters: config.Characters{Allowed: []string{"whitespace"}}}},
		{"period cannot end ASCII type", "abcABC123_", "trailing_period", config.TextPolicy{TrailingPeriod: "require"}},
		{"Latin required in ASCII type", "abcABC123_", "", config.TextPolicy{Characters: config.Characters{Required: []string{"Latin"}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			checker, err := Compile(tc.policy)
			if err != nil {
				t.Fatal(err)
			}
			err = checker.CheckAlphabet("a Conventional type", tc.alphabet)
			if tc.want == "" && err != nil || tc.want != "" && (err == nil || !strings.Contains(err.Error(), tc.want)) {
				t.Fatalf("CheckAlphabet() = %v, want %q", err, tc.want)
			}
		})
	}
}
