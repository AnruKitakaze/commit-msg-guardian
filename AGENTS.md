# Agent guidance

- This is a public Go CLI. Write user-facing documentation and examples in English.
- `main.go` owns the CLI; `config` loads YAML; `validator` applies policy; `internal/parser`, `internal/rules`, and `internal/gitrepo` handle messages, text rules, and Git.
- Project YAML drives both the local hook and CI. Build headers from type, scope, and literal separators; keep custom patterns as an escape hatch. Do not add hidden presets. Missing required fields need actionable errors; omitted optional fields add no extra checks.
- Keep old validation intent expressible in YAML, including the eleven Conventional types and length limits. Named text rules, old CLI flags, and the exported parser API are removed; document mappings and Unicode changes in `docs/migration-guide.md`.
- Local and CI validation both ignore lines starting with `#` and scissors content, even when Git stores those lines literally. CI range checks fail when the base, ref, or history is unavailable; when base and head resolve to the same commit, check that commit once.
- Keep `README.md` and `docs/configuration-examples.md` aligned with behavior. Make examples copyable and show accepted and rejected messages.
- For code changes, run `go test ./...`; always run `git diff --check`. Do not create commits unless explicitly asked.
