package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckRangeAndMissingConfig(t *testing.T) {
	repo := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, output)
		}
		return strings.TrimSpace(string(output))
	}
	git("init", "-q")
	git("config", "user.name", "Test User")
	git("config", "user.email", "test@example.com")
	git("-c", "core.hooksPath=/dev/null", "commit", "--allow-empty", "-q", "-m", "[TASK-1] base")
	base := git("rev-parse", "HEAD")
	git("-c", "core.hooksPath=/dev/null", "commit", "--allow-empty", "-q", "-m", "chore: bump dependency")
	good := git("rev-parse", "HEAD")
	git("-c", "core.hooksPath=/dev/null", "commit", "--allow-empty", "-q", "-m", "unrelated message")
	bad := git("rev-parse", "HEAD")

	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })
	var output bytes.Buffer
	if err := run([]string{"check-config"}, &output); err == nil || !strings.Contains(err.Error(), "create .commit-msg-guardian.yaml") {
		t.Fatalf("missing config error = %v", err)
	}
	configPath := filepath.Join(repo, ".commit-msg-guardian.yaml")
	contents := fmt.Sprintf("version: 1\ncommit:\n  formats:\n    - id: jira\n      scope:\n        brackets: square\n        issue:\n          projects: [TASK]\n    - id: chore\n      type:\n        allowed: [chore]\n      separators: [':']\nci:\n  commits:\n    base:\n      ref: %s\n", base)
	if err := os.WriteFile(configPath, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	output.Reset()
	err = run([]string{"check-range"}, &output)
	if err == nil || !strings.Contains(err.Error(), bad) {
		t.Fatalf("check-range error = %v, want bad commit %s", err, bad)
	}
	if !strings.Contains(output.String(), "checked 2 commit(s)") || !strings.Contains(output.String(), base) {
		t.Fatalf("check-range output = %q", output.String())
	}
	for _, tc := range []struct {
		name, sha string
		valid     bool
	}{
		{"valid same base and head", good, true},
		{"invalid same base and head", bad, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sameConfig := filepath.Join(repo, "same-"+tc.name+".yaml")
			if err := os.WriteFile(sameConfig, []byte(strings.Replace(contents, base, tc.sha, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			output.Reset()
			err := run([]string{"--config", sameConfig, "check-range", "--head", tc.sha}, &output)
			if (err == nil) != tc.valid {
				t.Fatalf("check-range error = %v, want valid %v", err, tc.valid)
			}
			if !strings.Contains(output.String(), "checking that commit") || !strings.Contains(output.String(), "checked 1 commit(s)") {
				t.Fatalf("same base and head output = %q", output.String())
			}
		})
	}
	messageFile := filepath.Join(repo, "message.txt")
	if err := os.WriteFile(messageFile, []byte("[TASK-1234] something\n# editor comment\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"check-file", messageFile}, &output); err != nil {
		t.Fatalf("check-file = %v", err)
	}

	commentConfig := filepath.Join(repo, "comment-policy.yaml")
	if err := os.WriteFile(commentConfig, []byte("version: 1\ncommit:\n  formats:\n    - id: chore\n      type:\n        allowed: [chore]\n      separators: [':']\n      body:\n        required: true\n"), 0600); err != nil {
		t.Fatal(err)
	}
	commentMessage := "chore: update dependency\n\n# literal body"
	git("-c", "core.hooksPath=/dev/null", "commit", "--allow-empty", "-q", "-m", commentMessage)
	commentCommit := git("rev-parse", "HEAD")
	if stored := git("log", "-1", "--pretty=%B"); !strings.Contains(stored, "# literal body") {
		t.Fatalf("Git did not store literal comment line: %q", stored)
	}
	if err := os.WriteFile(messageFile, []byte(commentMessage), 0600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"--config", commentConfig, "check-file", messageFile},
		{"--config", commentConfig, "check-commit", commentCommit},
	} {
		if err := run(args, &output); err == nil || !strings.Contains(err.Error(), "body is required") {
			t.Errorf("run(%v) = %v, want body requirement after comment stripping", args, err)
		}
	}
}
