# Commit Message Guardian

Commit Message Guardian validates Git commit messages locally through a `commit-msg` hook and in CI against a range of commits. Both paths read the same project-owned `.commit-msg-guardian.yaml` file.

## Configure a project

Add `.commit-msg-guardian.yaml` at the Git repository root, here is minimalistic example:

```yaml
# Conventional Commits, including `!:`
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

**OK:**

- `feat: add endpoint`
- `fix(api): correct response`
- `feat!: replace old API`
- `feat(api)!: replace old API`

**Not OK:**

- `custom: add endpoint` (type not listed)
- `feat[api]: add endpoint` (wrong brackets)
- `feat! : replace API` (wrong separator)

This is a small Conventional-style policy. A format is built from the blocks you include: `type`, `scope`, `separators`, `subject`, `body`, and `trailers`. For optional scopes, `!:` separators, Jira keys, renovate bot commits, CI bases, and a [complete field reference](docs/configuration-examples.md#complete-field-reference), see the [configuration examples](docs/configuration-examples.md).

*The examples are copyable YAML, not presets loaded by the tool.*

If your team uses both Jira task commits and taskless chores, configure two formats as shown in the [permanent policy example](docs/configuration-examples.md#permanent-policy-with-two-formats).

Use `subject.characters` to constrain writing systems; the examples include [Latin-only subjects](docs/configuration-examples.md#english-subject) and [Cyrillic subjects with English terms](docs/configuration-examples.md#cyrillic-subject-with-english-terms).

The config file and its required fields must be present. Every command validates the configuration before checking messages; a missing file or required field fails with a nonzero exit code and an error identifying the relevant configuration field or format. Unknown keys, character groups, contradictory settings, and invalid values also fail. Run `commit-msg-guardian check-config` after editing the file.

Optional fields do not add restrictions when omitted. A missing `type` or `scope` block means that component is absent from the header; an empty `type: {}` allows any syntactically valid type. `subject` is the only component every format has. See [required fields and omitted checks](docs/configuration-examples.md#required-fields-and-omitted-checks) for details.

## Use the local hook

Add this repository to `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/AnruKitakaze/commit-msg-guardian
    rev: <major-release-tag>
    hooks:
      - id: commit-msg-guardian
```

Pin `rev` to a published release tag, then run `pre-commit install --hook-type commit-msg`. The hook calls `commit-msg-guardian check-file <message-file>`.

## Check commits in CI

Add a base to the same project file:

```yaml
ci:
  commits:
    base:
      ref: origin/main
```

Install the binary in the CI job, fetch full Git history and the target ref, then run:

```sh
commit-msg-guardian check-range
```

`ci.commits.base.ref` can be a fixed ref such as `origin/main` or `origin/release/something`. For a target branch chosen by the CI job, configure `ci.commits.base.env: COMMIT_MSG_GUARDIAN_BASE_REF` and provide that variable in the job. Exactly one of `ref` and `env` is allowed. `--head <ref>` selects the source head when CI checks out a synthetic merge commit; otherwise `HEAD` is used.

The command prints the resolved base and head SHA, then checks every commit reachable from head but not from base. If both refs resolve to the same commit, it checks that commit once, which also covers a direct commit to the target branch. Missing refs, missing environment variables, invalid configs, and invalid messages fail the command. A CI platform can mark this job optional; for example GitLab uses `allow_failure: true`. The validator still exits with an error so the failed check remains visible.

Additional commands:

```sh
commit-msg-guardian check-file <message-file>
commit-msg-guardian check-commit <sha-or-ref>
commit-msg-guardian check-config
commit-msg-guardian --config path/to/policy.yaml check-config
```

All commands ignore lines starting with `#` and discard the scissors line and everything after it. This is Commit Message Guardian's rule even if Git stored those lines literally, for example with `git commit -m`.

## Migrate from v0.x

The [migration guide](docs/migration-guide.md) maps every old flag to YAML, shows the old default policy, and covers hook and CI changes. It includes a ready-to-use prompt for an AI assistant. The old CLI flags and Go parser API were removed.

The executable remains at the module root, so `go install github.com/AnruKitakaze/commit-msg-guardian@<version>` keeps its existing path. `config` and `validator` are importable packages; `internal` packages are implementation details.
