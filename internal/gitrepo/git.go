package gitrepo

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var authorEmail = regexp.MustCompile(`<([^<>]+)>`)

func run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %s: %w", strings.Join(args, " "), strings.TrimSpace(string(output)), err)
	}
	return strings.TrimSpace(string(output)), nil
}

func Root() (string, error) { return run("rev-parse", "--show-toplevel") }

func AuthorEmail() (string, error) {
	identity, err := run("var", "GIT_AUTHOR_IDENT")
	if err != nil {
		return "", err
	}
	match := authorEmail.FindStringSubmatch(identity)
	if match == nil {
		return "", fmt.Errorf("cannot read author email from Git identity")
	}
	return match[1], nil
}

func Resolve(ref string) (string, error) {
	if ref == "" {
		return "", fmt.Errorf("empty Git ref")
	}
	return run("rev-parse", "--verify", "--end-of-options", ref+"^{commit}")
}

func Range(base, head string) ([]string, error) {
	shallow, err := run("rev-parse", "--is-shallow-repository")
	if err != nil {
		return nil, err
	}
	if shallow == "true" {
		return nil, fmt.Errorf("repository history is shallow; fetch full history before check-range")
	}
	output, err := run("rev-list", "--reverse", base+".."+head)
	if err != nil || output == "" {
		return nil, err
	}
	return strings.Split(output, "\n"), nil
}

// Commit reads a stored message; the validator then applies the project's
// comment and scissors convention to the stored text.
func Commit(sha string) (message, email string, err error) {
	cmd := exec.Command("git", "show", "-s", "--format=%ae%x00%B", sha)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("read commit %s: %s: %w", sha, strings.TrimSpace(string(output)), err)
	}
	parts := bytes.SplitN(output, []byte{0}, 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("read commit %s: missing author or message", sha)
	}
	return string(parts[1]), strings.TrimSpace(string(parts[0])), nil
}
