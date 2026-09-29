# Configuration examples

Copy a policy into `.commit-msg-guardian.yaml` at your Git repository root, then run `commit-msg-guardian check-config`. These are examples, not presets that the tool downloads or changes later. The local hook and CI read the same file.

## Required fields and omitted checks

Missing required fields fail at startup with an error pointing to the field or format. Omitting an optional field adds no extra check.

**Required fields**

- `version: 1`, at least one `commit.formats` entry, and an `id` in each entry.
- A `pattern`, when used, must cover the whole header (`^...$`) and contain a named `subject` group. A configured `type` or `scope` needs the matching named group.
- `when.author_email` if `when` is present; `name` and `value_pattern` for every `trailers.required` entry.
- `body.optional_for_types` needs `type`. `separators` needs `type` or `scope`. If both `type` and `scope` are present, use round or square scope brackets; `scope.required: false` also needs `type`.
- One of `ci.commits.base.ref` or `ci.commits.base.env` for `check-range` only.

**Optional fields**

- `type` and `scope`: if absent, that component does not appear in the header. An empty `type: {}` accepts any syntactically valid type. A configured scope is required unless `scope.required: false` is set.
- `scope.brackets`: `none` by default; choose `round` or `square` to wrap the scope. `separators`: literal endings before the required space and subject; omit it for no punctuation.
- `type.allowed`, `scope.allowed`, `scope.issue.projects`, `characters`, `structure`, `case`, length and line limits, `body`, `trailers`, and `when`: omitted restrictions do not run. `scope.issue: {}` accepts any syntactically valid issue key.

A subject is always required and cannot be empty. Body checks still run on an empty body if a positive minimum length or required character group is configured; `body.optional_for_types` exempts its listed types when their body is empty. A plain subject-only format accepts nearly any nonempty header, so review it before combining it with stricter general formats.

## Complete field reference

This copyable configuration shows every supported field. The first format uses the constructor. The second shows the regex escape hatch, which cannot share `separators`, `scope.brackets`, or `scope.issue` with the constructor.

```yaml
version: 1
commit:
  formats:
    - id: typed-task
      when:
        author_email: developer@example.com
      type:
        allowed: [feat, docs]
        characters:
          allowed: [Latin]
          required: [Latin]
          forbidden: [Cyrillic]
      scope:
        required: true
        brackets: round             # Alternatives: square, none
        issue:
          projects: [TASK]
        allowed: [TASK-1234]
        structure: flat             # feat(TASK-1234): ...
        # structure: slash-separated # For api/client-v2 after removing issue and allowed
        characters:
          allowed: [Latin, digit, punctuation]
          required: [Latin, digit]
      separators: [":", "!:"]       # feat(TASK-1234)!: ...
      subject:
        min_length: 1
        max_length: 72
        max_lines: 1
        case: lower-first-letter    # Alternative: upper-first-letter
        trailing_period: forbid     # Alternative: require
        characters:
          allowed: [Latin, digit, whitespace, punctuation]
          required: [Latin]
          forbidden: [Cyrillic]
      body:
        required: true
        optional_for_types: [docs]
        min_length: 0
        max_line_length: 100
        max_total_length: 2000
        max_lines: 20
        trailing_period: forbid
        characters:
          allowed: [Latin, digit, whitespace, punctuation]
      trailers:
        required:
          - name: Refs
            value_pattern: '^TASK-[1-9][0-9]*$'

    - id: custom-header
      pattern: '^TASK-(?P<issue>[1-9][0-9]*) :: (?P<subject>.+)$'
      subject:
        max_length: 72

ci:
  commits:
    base:
      ref: origin/main
      # env: COMMIT_MSG_GUARDIAN_BASE_REF # Alternative to ref; never both.
```

- `characters.allowed` is a union. Every group in `characters.required` must occur; any matching group in `characters.forbidden` rejects the text. Without `allowed`, all groups are allowed.
- Character groups include Unicode scripts such as `Latin` and `Cyrillic`, plus `letter`, `digit`, `whitespace`, `punctuation`, `symbol`, `mark`, and `other`. They check writing systems, not natural language. Include `symbol` for `+` in `C++`.
- `structure: flat` accepts one scope segment; `slash-separated` accepts segments joined by `/`. Each segment uses Latin letters, digits, and hyphens and starts and ends with a letter or digit.
- Lengths count Unicode code points. `body` limits apply to body text only; a final trailer block is checked separately.
- `when.author_email` overrides all general formats for that author. CI reads the stored commit author; the local hook uses the Git author identity.

## Conventional Commits, including `!:`

```yaml
version: 1
commit:
  formats:
    - id: conventional
      type:
        allowed: [feat, fix, docs, chore]
      scope:
        brackets: round
        required: false
      separators: [":", "!:"]
```

- **OK:** `feat: add endpoint`; `fix(api): correct response`; `feat!: replace old API`; `feat(api)!: replace old API`.
- **Not OK:** `custom: add endpoint` (type not listed); `feat[api]: add endpoint` (wrong brackets); `feat! : replace API` (wrong separator).

`!:` is a literal separator option. It permits the Conventional Commits breaking-change marker after the type or scope. A `BREAKING CHANGE: ...` footer is also allowed; neither form requires the other. The [specification](https://www.conventionalcommits.org/en/v1.0.0/) does not prescribe a type list.

## Jira key in square brackets

```yaml
version: 1
commit:
  formats:
    - id: jira-task
      scope:
        brackets: square
        issue:
          projects: [TASK]
```

- **OK:** `[TASK-1234] do something` (body optional).
- **Not OK:** `(TASK-1234) do something` (wrong brackets); `[OTHER-1234] do something` (wrong project); `[TASK-1234]` (empty subject).

The issue key is checked locally; the tool does not contact Jira. To allow round brackets too, add a second format with `scope.brackets: round`. For `TASK-1234: subject`, use `scope.brackets: none` and `separators: [":"]`.

## Type with a Jira task in scope

```yaml
version: 1
commit:
  formats:
    - id: typed-task
      type:
        allowed: [feat, fix]
      scope:
        brackets: round
        issue:
          projects: [TASK]
      separators: [":"]
```

- **OK:** `feat(TASK-1234): add feature`.
- **Not OK:** `feat(OTHER-1234): add feature` (wrong project); `feat(api): add feature` (not an issue key).

## English subject

Add this to a format's `subject` block:

```yaml
subject:
  characters:
    allowed: [Latin, digit, whitespace, punctuation]
    required: [Latin]
```

- **OK:** `fix: update API v2` in a format with `type` and `separators: [":"]`.
- **Not OK:** `fix: обновить API` (Cyrillic); `fix: 123` (no Latin letter); `fix: update C++ bindings` (`+` is a symbol).

## Cyrillic subject with English terms

Add this to a format's `subject` block:

```yaml
subject:
  characters:
    allowed: [Cyrillic, Latin, digit, whitespace, punctuation]
    required: [Cyrillic]
```

- **OK:** `[TASK-1234] Исправить API flow v2` in the square-bracket Jira format above.
- **Not OK:** `[TASK-1234] Fix API flow` (no Cyrillic letter); `[TASK-1234] 改善 API flow` (other script).

This detects an all-Latin subject; it cannot prove that the sentence is Russian. Multiple entries in `required` mean **each** group must occur.

## Angular-style policy

```yaml
version: 1
commit:
  formats:
    - id: angular-like
      type:
        allowed: [build, ci, docs, feat, fix, perf, refactor, test]
      scope:
        brackets: round
        required: false
        allowed: [api, core, docs, ui]
      separators: [":"]
      subject:
        case: lower-first-letter
        trailing_period: forbid
      body:
        required: true
        optional_for_types: [docs]
        min_length: 20
```

- **OK:** `docs: correct examples` (body exempt); `feat(api): add pagination` followed by a blank line and at least 20 body characters.
- **Not OK:** `feat(api): add pagination` without a body; `feat(api): Add pagination.` (case and period).

This is an illustrative policy inspired by [Angular's guidelines](https://github.com/angular/angular/blob/main/contributing-docs/commit-message-guidelines.md).

## Plain Git subject

```yaml
version: 1
commit:
  formats:
    - id: plain
      subject:
        max_length: 50
        case: upper-first-letter
      body:
        max_line_length: 72
```

- **OK:** `Improve error messages`.
- **Not OK:** `improve error messages` (first letter must be uppercase).

## Permanent policy with two formats

```yaml
version: 1
commit:
  formats:
    - id: jira-task
      scope:
        brackets: square
        issue:
          projects: [TASK]
    - id: taskless-chore
      type:
        allowed: [chore]
      scope:
        brackets: round
        required: false
      separators: [":"]
```

- **OK:** `[TASK-1234] fix cache`; `chore: bump packageX version`; `chore(deps): update packageX`.
- **Not OK:** `feat: add cache` (type not allowed); `[TASK-1234]` (empty subject).

Branch names do not select a format. Both formats form the permanent policy.

## Renovate as an author-specific format

Add a third format to the two above. Replace the email with the **Git commit author email** in your repository; a merge request account name does not establish that value.

```yaml
    - id: renovate-update
      when:
        author_email: renovate-bot@example.invalid
      type:
        allowed: [chore, fix]
      scope:
        brackets: round
        allowed: [deps]
      separators: [":"]
```

- **OK for the configured author:** `fix(deps): update module google.golang.org/grpc to v1.23.4 [security]`; `chore(deps): update module golang.org/x/text to v0.12.3 [security]`.
- **Not OK for that author:** `chore: update packageX` (scope required), even though a general format accepts it for other authors.

Renovate can use both `chore(deps)` and `fix(deps)` ([semantic commits](https://docs.renovatebot.com/semantic-commits/)). Omit `when` if everyone may use the format.

## Required trailer and custom header

A required final trailer can be added to any format:

```yaml
trailers:
  required:
    - name: Refs
      value_pattern: '^TASK-[1-9][0-9]*$'
```

- **OK:** `feat(api): add pagination` followed by a blank line and `Refs: TASK-1234` at the end.
- **Not OK:** The same message without `Refs: TASK-1234`, or with `Refs: OTHER-1234`.

Required trailer checks recognize single-line entries in the final trailer block. A footer follows the body, or follows the subject after a blank line when there is no body.

For a header the constructor cannot express, use `pattern`:

```yaml
version: 1
commit:
  formats:
    - id: custom-task
      pattern: '^TASK-(?P<issue>[1-9][0-9]*) :: (?P<subject>.+)$'
      subject:
        max_length: 72
```

- **OK:** `TASK-123 :: improve output`.
- **Not OK:** `TASK-123: improve output` (wrong punctuation); `TASK-123 :: ` (empty subject).

A pattern must match the entire header and include named `subject`. Optional named groups are `type`, `scope`, and `issue`; `type` and `scope` can then have their usual text checks. `pattern` and `separators` are mutually exclusive.

## CI base and checked commits

Add one base source to a full project configuration:

```yaml
ci:
  commits:
    base:
      ref: origin/release/something
```

For a target chosen by CI, replace `ref` with an environment variable name:

```yaml
ci:
  commits:
    base:
      env: COMMIT_MSG_GUARDIAN_BASE_REF
```

- Set `COMMIT_MSG_GUARDIAN_BASE_REF=origin/release/something` when using `env`. `check-range` checks commits in `<base>..<head>`; `HEAD` is the default head. See [Git revision ranges](https://git-scm.com/docs/git-rev-parse#_specifying_ranges).
- When base and head resolve to the same commit, that commit is checked once. Other empty ranges check zero commits and succeed.
- Fetch the target ref and full history. A missing ref, variable, or history fails explicitly. In GitLab, `GIT_DEPTH: "0"` disables shallow cloning, but the target ref must still be available ([GitLab docs](https://docs.gitlab.com/ci/runners/configure_runners/#shallow-cloning)).
- For a synthetic merge checkout, pass `check-range --head <source-ref>` so the source commit is checked.
- Both local and CI checks ignore lines starting with `#` and content from a scissors line onward, even if Git stored them literally. Indented `#` lines remain content. `pre-commit run --all-files` does not validate historical commit messages ([pre-commit docs](https://pre-commit.com/)).
- `fixup!`, `squash!`, `amend!`, merge, and revert commits have no implicit exceptions. A platform-created final squash or merge commit needs a separate check if your policy covers it.
- An invalid config fails both local and CI commands. An optional CI job can keep a visible warning, for example GitLab's [`allow_failure: true`](https://docs.gitlab.com/ci/yaml/#allow_failure).
