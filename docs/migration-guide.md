# Migrate from v0.x to v1

The hook now reads `.commit-msg-guardian.yaml` from the Git repository root. Old CLI flags and named text rules are gone. Migrate the behavior you need, then add new formats.

## 1. Create the project policy

This maps the old defaults to the new policy. Copy it to `.commit-msg-guardian.yaml` and review the Unicode differences below:

```yaml
version: 1
commit:
  formats:
    - id: conventional
      type:
        allowed: [feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert]
        characters:
          allowed: [Latin, digit, whitespace, punctuation]
      scope:
        brackets: round
        required: false
        structure: flat              # feat(api-v2): ...
        # structure: slash-separated # feat(api/client-v2): ...
      separators: [":", "!:"]       # The old parser accepted feat!: ...
      subject:
        characters:
          forbidden: [Cyrillic]
```

If your hook set custom flags, move each rule to the corresponding field:

| Old flag | YAML location |
| --- | --- |
| `--type-rules` | `type` |
| `--scope-rules` | `scope` |
| `--description-rules` | `subject` |
| `--body-rules` | `body` |
| `--description-length-limit` | `subject.max_length` |
| `--body-length-limit` | `body.max_total_length` |

Within that field, replace each old rule as follows:

| Old rule | New field setting |
| --- | --- |
| `noCyrillic` | `characters.forbidden: [Cyrillic]` |
| `noLatin` | `characters.forbidden: [Latin]` |
| `noDigits` | `characters.forbidden: [digit]` |
| `cyrillicOnly` | `characters.allowed: [Cyrillic, whitespace, punctuation]` |
| `latinOnly` | `characters.allowed: [Latin, whitespace, punctuation]` |
| `digitsOnly` | `characters.allowed: [digit, whitespace, punctuation]` |
| `allowLatin`, `allowDigits` | `characters.allowed: [Latin, digit, whitespace, punctuation]` |
| `allowCyrillic` | `characters.allowed: [Cyrillic, digit, whitespace, punctuation]` |
| `allowScope` | `structure: flat` |
| `allowPathScope` | `structure: slash-separated` |
| `capitalized` | `case: upper-first-letter` |
| `oneLine` | `max_lines: 1` |
| `trailingPeriod` | `trailing_period: require` |
| `noTrailingPeriod` | `trailing_period: forbid` |

Keep `type.allowed` explicit: the old eleven-type list is **not** a new default. An empty old rule flag means omit its new settings; a length limit of `0` means omit the limit. Multiple old rules were combined with **AND**: intersect their allowed groups instead of merging them. For example, `allowLatin,allowCyrillic` allowed digits, whitespace, and punctuation but no letters; use `characters.allowed: [digit, whitespace, punctuation]` if that was intentional.

The header now uses a constructor instead of `kind`. `type: {}` enables an unrestricted type; `scope.brackets: round` with `required: false` makes the scope optional. `separators: [":", "!:"]` keeps ordinary and breaking-change headers. To migrate a Jira prefix, use `scope.issue.projects` plus `scope.brackets: square` (or `round`), as shown in the [examples](configuration-examples.md#jira-key-in-square-brackets). A custom header can still use an anchored `pattern` with a named `subject` group.

The new groups use Unicode scripts and character categories. Old `noLatin`/`latinOnly`/`allowLatin` checked ASCII A–Z, while `Latin` now includes accented letters. `digit` and `whitespace` also cover Unicode characters beyond the old regular expressions. Script groups here classify letters; combining marks require `mark`, whereas the old Cyrillic regex could match Cyrillic marks. Symbols such as `+` need `symbol`, even when `punctuation` is allowed. Review representative messages if exact old behavior matters.

Final trailers are now parsed separately from the body. Body requirements and limits apply to body text, not a `Refs:` or `BREAKING CHANGE:` trailer. Test messages with trailers if you used body rules before upgrading.

For Jira, bot, or other new policies, use the [configuration examples](configuration-examples.md) after the old behavior is covered.

## 2. Update the hook

Pin the new major release in `.pre-commit-config.yaml` and remove the old `args`:

```yaml
repos:
  - repo: https://github.com/AnruKitakaze/commit-msg-guardian
    rev: <v1-release-tag>
    hooks:
      - id: commit-msg-guardian
```

Run `pre-commit install --hook-type commit-msg`. If you invoke the binary directly, replace `commit-msg-guardian [old flags] <message-file>` with `commit-msg-guardian check-file <message-file>`; the hook definition already makes that change for pre-commit users.

## 3. Add CI validation if needed

`check-range` is new. To use it, add one base source to the policy:

```yaml
ci:
  commits:
    base:
      ref: origin/main
```

Use `env: COMMIT_MSG_GUARDIAN_BASE_REF` instead when your CI job chooses the target branch. Fetch the target ref and full history, then run `commit-msg-guardian check-range`. The [CI examples](configuration-examples.md#ci-base-and-checked-commits) cover release branches and checkout details. No `ci` block is needed for the local hook.

## 4. Verify the result

- Run `commit-msg-guardian check-config` to catch missing fields, unknown keys, and misspelled rules.
- Run `commit-msg-guardian check-file <message-file>` with messages your old hook accepted and rejected. Confirm both outcomes still match.
- If using CI, run `commit-msg-guardian check-range` in a checkout with the target ref and full history.

## Let an AI assistant do the migration

Give it this guide and your repository, then use this prompt:

```text
Migrate this repository from Commit Message Guardian v0.x to v1 using this guide.
Read the existing .pre-commit-config.yaml, hook arguments, and any CI job that invokes the tool. Do not read secret files.
Create .commit-msg-guardian.yaml at the Git repository root. Preserve the old eleven-type allowlist, effective rule combinations, length limits, and hook stage, including custom overrides. Translate every named rule using the mapping above; do not leave old rule names or checks lists in the new YAML. Do not silently broaden or narrow the existing policy.
Update the pre-commit hook to the new major release and remove obsolete CLI arguments. Update direct invocations to use check-file. If CI already checks commit messages, configure check-range with the actual target ref or environment variable; do not guess the base ref.
Run check-config and test representative previously accepted and rejected messages. Report Unicode-related behavior changes, assumptions, and validation results. Leave the changes for review without committing them.
```

Go integrations also change: `parser.ParseCommitMessage` and `rules.ConventionalCommitTypes` were removed. Use `config.Load`, `validator.New`, and `Validator.Validate` for direct validation.
