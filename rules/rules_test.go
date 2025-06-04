package rules

import (
	"testing"
)

func TestRuleFactory(t *testing.T) {
	tests := []struct {
		name    string
		rule    string
		wantErr bool
	}{
		{"valid nocyrillic", "nocyrillic", false},
		{"valid nolatin", "nolatin", false},
		{"valid nodigits", "nodigits", false},
		{"valid cyrilliconly", "cyrilliconly", false},
		{"valid latinonly", "latinonly", false},
		{"valid digitsonly", "digitsonly", false},
		{"valid allowlatin", "allowlatin", false},
		{"valid allowcyrillic", "allowcyrillic", false},
		{"valid allowdigits", "allowdigits", false},
		{"valid allowscope", "allowscope", false},
		{"invalid rule", "nonexistent", true},
		// Case insensitivity tests
		{"uppercase rule", "NOCYRILLIC", false},
		{"mixed case rule", "NoLaTiN", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := RuleFactory(tt.rule)
			if (err != nil) != tt.wantErr {
				t.Errorf("RuleFactory() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNoCyrillicRule(t *testing.T) {
	rule := &NoCyrillicRule{}
	tests := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{"latin only", "Hello world", false},
		{"with digits", "Hello 123", false},
		{"with punctuation", "Hello, world!", false},
		{"with cyrillic", "Hello привет", true},
		{"cyrillic only", "привет", true},
		{"mixed", "Hello привет 123", true},
	}

	runRuleTests(t, "NoCyrillicRule", rule, tests)
}

func TestNoLatinRule(t *testing.T) {
	rule := &NoLatinRule{}
	tests := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{"cyrillic only", "привет", false},
		{"with digits", "привет 123", false},
		{"with punctuation", "привет, мир!", false},
		{"with latin", "привет hello", true},
		{"latin only", "hello", true},
		{"mixed", "hello привет 123", true},
	}

	runRuleTests(t, "NoLatinRule", rule, tests)
}

func TestNoDigitsRule(t *testing.T) {
	rule := &NoDigitsRule{}
	tests := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{"latin only", "Hello world", false},
		{"cyrillic only", "привет мир", false},
		{"with punctuation", "Hello, world!", false},
		{"with digits", "Hello 123", true},
		{"digits only", "123", true},
		{"mixed", "Hello 123 привет", true},
	}

	runRuleTests(t, "NoDigitsRule", rule, tests)
}

func TestCyrillicOnlyRule(t *testing.T) {
	rule := &CyrillicOnlyRule{}
	tests := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{"cyrillic only", "привет", false},
		{"cyrillic with spaces", "привет мир", false},
		{"cyrillic with punctuation", "привет, мир!", false},
		{"with latin", "привет hello", true},
		{"with digits", "привет 123", true},
		{"mixed", "hello привет 123", true},
	}

	runRuleTests(t, "CyrillicOnlyRule", rule, tests)
}

func TestLatinOnlyRule(t *testing.T) {
	rule := &LatinOnlyRule{}
	tests := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{"latin only", "Hello", false},
		{"latin with spaces", "Hello world", false},
		{"latin with punctuation", "Hello, world!", false},
		{"with cyrillic", "Hello привет", true},
		{"with digits", "Hello 123", true},
		{"mixed", "Hello привет 123", true},
	}

	runRuleTests(t, "LatinOnlyRule", rule, tests)
}

func TestDigitsOnlyRule(t *testing.T) {
	rule := &DigitsOnlyRule{}
	tests := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{"digits only", "123", false},
		{"digits with spaces", "123 456", false},
		{"digits with punctuation", "123, 456!", false},
		{"with latin", "123 hello", true},
		{"with cyrillic", "123 привет", true},
		{"mixed", "hello 123 привет", true},
	}

	runRuleTests(t, "DigitsOnlyRule", rule, tests)
}

func TestAllowLatinRule(t *testing.T) {
	rule := &AllowLatinRule{}
	tests := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{"latin only", "Hello", false},
		{"latin with spaces", "Hello world", false},
		{"latin with punctuation", "Hello, world!", false},
		{"latin with digits", "Hello 123", false},
		{"latin with hyphen", "Hello-world", false},
		{"digits only", "123", false},
		{"with cyrillic", "Hello привет", true},
		{"cyrillic only", "привет", true},
	}

	runRuleTests(t, "AllowLatinRule", rule, tests)
}

func TestAllowCyrillicRule(t *testing.T) {
	rule := &AllowCyrillicRule{}
	tests := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{"cyrillic only", "привет", false},
		{"cyrillic with spaces", "привет мир", false},
		{"cyrillic with punctuation", "привет, мир!", false},
		{"cyrillic with digits", "привет 123", false},
		{"cyrillic with hyphen", "привет-мир", false},
		{"digits only", "123", false},
		{"with latin", "привет hello", true},
		{"latin only", "hello", true},
	}

	runRuleTests(t, "AllowCyrillicRule", rule, tests)
}

func TestAllowDigitsRule(t *testing.T) {
	rule := &AllowDigitsRule{}
	tests := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{"digits only", "123", false},
		{"digits with spaces", "123 456", false},
		{"digits with punctuation", "123, 456!", false},
		{"digits with latin", "hello 123", false},
		{"latin only", "hello", false},
		{"with hyphen", "hello-123", false},
		{"with cyrillic", "123 привет", true},
		{"cyrillic only", "привет", true},
	}

	runRuleTests(t, "AllowDigitsRule", rule, tests)
}

func TestAllowScopeRule(t *testing.T) {
	rule := &AllowScopeRule{}
	tests := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{"simple scope", "T1", false},
		{"scope with hyphen", "T-1", false},
		{"complex scope", "task-123", false},
		{"multiple hyphens", "my-task-123", false},
		{"starts with letter", "a-1", false},
		{"starts with digit", "1-a", false},
		{"empty string", "", true},
		{"single char", "T", false},
		{"starts with hyphen", "-T1", true},
		{"ends with hyphen", "T1-", true},
		{"only hyphens", "-", true},
		{"with cyrillic", "task-привет", true},
		{"with spaces", "task 123", true},
		{"with other punctuation", "task.123", true},
	}

	runRuleTests(t, "AllowScopeRule", rule, tests)
}

// Helper function to run rule tests
func runRuleTests(t *testing.T, ruleName string, rule Rule, tests []struct {
	name    string
	text    string
	wantErr bool
}) {
	for _, tt := range tests {
		t.Run(ruleName+"/"+tt.name, func(t *testing.T) {
			err := rule.Validate(tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
