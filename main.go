package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/AnruKitakaze/commit-msg-guardian/config"
	"github.com/AnruKitakaze/commit-msg-guardian/internal/gitrepo"
	"github.com/AnruKitakaze/commit-msg-guardian/validator"
)

const usage = `Usage: commit-msg-guardian [--config path] <command>

Commands:
  check-file <path>     Validate a draft commit message (local commit-msg hook)
  check-commit <ref>    Validate one stored Git commit
  check-range [--head ref]  Validate commits outside ci.commits.base
  check-config         Validate the project YAML file
`

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "commit-msg-guardian:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		_, _ = io.WriteString(stdout, usage)
		return nil
	}
	flags := flag.NewFlagSet("commit-msg-guardian", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	configPath := flags.String("config", "", "path to project YAML configuration")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, _ = io.WriteString(stdout, usage)
			return nil
		}
		return err
	}
	args = flags.Args()
	if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
		_, _ = io.WriteString(stdout, usage)
		return nil
	}
	if len(args) == 0 {
		return fmt.Errorf("%s", usage)
	}
	switch args[0] {
	case "check-config", "check-file", "check-commit", "check-range":
	default:
		return fmt.Errorf("unknown command %q\n%s", args[0], usage)
	}
	path := *configPath
	if path == "" {
		root, err := gitrepo.Root()
		if err != nil {
			root, err = os.Getwd()
			if err != nil {
				return err
			}
		}
		path = filepath.Join(root, ".commit-msg-guardian.yaml")
	}
	cfg, err := config.Load(path)
	if err != nil {
		return fmt.Errorf("%w (create .commit-msg-guardian.yaml or pass --config)", err)
	}
	v, err := validator.New(cfg)
	if err != nil {
		return fmt.Errorf("invalid config %q: %w", path, err)
	}
	switch args[0] {
	case "check-config":
		if len(args) != 1 {
			return fmt.Errorf("check-config takes no arguments")
		}
		fmt.Fprintf(stdout, "config OK: %s\n", path)
		ids := make([]string, 0, len(cfg.Commit.Formats))
		for _, f := range cfg.Commit.Formats {
			ids = append(ids, f.ID)
		}
		fmt.Fprintf(stdout, "formats: %s\n", strings.Join(ids, ", "))
		return nil
	case "check-file":
		if len(args) != 2 {
			return fmt.Errorf("check-file requires one message file path")
		}
		message, err := os.ReadFile(args[1])
		if err != nil {
			return fmt.Errorf("read message file %q: %w", args[1], err)
		}
		email := ""
		if v.RequiresAuthor() {
			email, err = gitrepo.AuthorEmail()
			if err != nil {
				return err
			}
		}
		return v.Validate(string(message), email)
	case "check-commit":
		if len(args) != 2 {
			return fmt.Errorf("check-commit requires one commit ref")
		}
		sha, err := gitrepo.Resolve(args[1])
		if err != nil {
			return err
		}
		message, email, err := gitrepo.Commit(sha)
		if err != nil {
			return err
		}
		if err := v.Validate(message, email); err != nil {
			return fmt.Errorf("commit %s: %w", sha, err)
		}
		return nil
	case "check-range":
		return checkRange(args[1:], v, stdout)
	}
	return nil
}

func checkRange(args []string, v *validator.Validator, stdout io.Writer) error {
	flags := flag.NewFlagSet("check-range", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	headRef := flags.String("head", "HEAD", "source branch head")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("check-range accepts only --head")
	}
	base := v.Base()
	baseRef := base.Ref
	if base.Env != "" {
		baseRef = os.Getenv(base.Env)
		if baseRef == "" {
			return fmt.Errorf("ci.commits.base environment variable %q is empty or unset", base.Env)
		}
	}
	if baseRef == "" {
		return fmt.Errorf("ci.commits.base.ref or ci.commits.base.env is required for check-range")
	}
	baseSHA, err := gitrepo.Resolve(baseRef)
	if err != nil {
		return fmt.Errorf("resolve base %q: %w", baseRef, err)
	}
	headSHA, err := gitrepo.Resolve(*headRef)
	if err != nil {
		return fmt.Errorf("resolve head %q: %w", *headRef, err)
	}
	fmt.Fprintf(stdout, "checking commits: base %s (%s), head %s (%s)\n", baseRef, baseSHA, *headRef, headSHA)
	shas, err := gitrepo.Range(baseSHA, headSHA)
	if err != nil {
		return err
	}
	if len(shas) == 0 && baseSHA == headSHA {
		fmt.Fprintln(stdout, "base and head resolve to the same commit; checking that commit")
		shas = []string{headSHA}
	}
	var failures []error
	for _, sha := range shas {
		message, email, err := gitrepo.Commit(sha)
		if err == nil {
			err = v.Validate(message, email)
		}
		if err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", sha, err))
		}
	}
	fmt.Fprintf(stdout, "checked %d commit(s)\n", len(shas))
	if len(failures) > 0 {
		return fmt.Errorf("%d commit(s) failed:\n%w", len(failures), errors.Join(failures...))
	}
	return nil
}
