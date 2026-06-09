# Commit Message Guardian

A Git pre-commit hook that validates commit messages against specified rules, ensuring consistent message formatting and language usage.

## Features

- Validates commit messages against the [Conventional Commits](https://www.conventionalcommits.org/) format
- Supports various text validation rules:
  - Restrictive rules (allow ONLY specified characters):
    - `noCyrillic`: Prevents Cyrillic characters
    - `noLatin`: Prevents Latin characters
    - `noDigits`: Prevents digits
    - `cyrillicOnly`: Allows only Cyrillic characters
    - `latinOnly`: Allows only Latin characters
    - `digitsOnly`: Allows only digits
  - Permissive rules (ALLOW but don't require):
    - `allowLatin`: Allows Latin characters, digits, and basic punctuation
    - `allowCyrillic`: Allows Cyrillic characters, digits, and basic punctuation
    - `allowDigits`: Allows Latin characters, digits, and basic punctuation
    - `allowScope`: Special rule for scopes that allows Latin, digits, and hyphens (must start and end with alphanumeric)
    - `allowPathScope`: Special rule for scopes that allows slash-delimited `allowScope` path segments
  - Summary/body rules:
    - `capitalized`: Requires the first letter to be uppercase
    - `oneLine`: Requires text to stay on a single line
- Configurable validation rules for different parts of the commit message (type, scope, description, body)
- Configurable strict length limits for description and body
- Body text is not validated by default, but can be validated with `--body-rules`
- Supports breaking-change headers such as `feat!: Summary` and `feat(scope)!: Summary`

## Installation

To use this hook in your project:

1. Install [pre-commit](https://pre-commit.com/) if you haven't already:
```bash
pip install pre-commit
```

2. Add this to your `.pre-commit-config.yaml`:
```yaml
repos:
- repo: https://github.com/AnruKitakaze/commit-msg-guardian
  rev: v0.1.6  # Use the latest version
  hooks:
    - id: commit-msg-guardian
      # Optional: override default rules
      args:
        - --type-rules=allowLatin
        - --scope-rules=allowScope
        - --description-rules=noCyrillic,capitalized
        - --body-rules=oneLine
        - --description-length-limit=60
        - --body-length-limit=72
```

3. Install the commit-msg hook:
```bash
pre-commit install --hook-type commit-msg
```

**Note**: The standard `pre-commit install` command won't work for this hook as it's a commit-msg hook, not a pre-commit hook. Make sure to use the command above.

## Usage

The hook validates commit messages against the following format:
```
type(scope): description

[optional body]
```

Breaking-change headers are also allowed:
```
type!: description
type(scope)!: description
```

### Valid Commit Types

The following commit types are supported:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or modifying tests
- `build`: Build system changes
- `ci`: CI configuration changes
- `chore`: General maintenance
- `revert`: Reverting changes

### Command Line Arguments

You can customize the validation rules using command line arguments:

- `--type-rules`: Comma-separated rules for commit type (default: "allowLatin")
- `--scope-rules`: Comma-separated rules for commit scope (default: "allowScope")
- `--description-rules`: Comma-separated rules for commit description (default: "noCyrillic")
- `--body-rules`: Comma-separated rules for commit body (default: "")
- `--description-length-limit`: Strict description length limit; `0` disables the limit (default: 0)
- `--body-length-limit`: Strict body length limit; `0` disables the limit (default: 0)

Use `--scope-rules=allowPathScope` to allow slash-delimited scopes such as `app/api` or `this/is/some/path`.

### Examples

Valid commit messages:
```
feat(TGK-1827): This is an example
docs(T-1): This is valid too
feat(T1): Another valid example
feat(task-123): Valid with hyphen
feat(app/api): Valid with slash-delimited folders when using --scope-rules=allowPathScope
feat!: Breaking change without scope
feat(api)!: Breaking change with scope

With кириллица in description (body is not validated by default)
```

Invalid commit messages:
```
feat(TGK-1827): Забыл убрать кириллицу   # Contains Cyrillic in description
random: Not a valid type                 # Invalid commit type
feat[scope]: Wrong scope format          # Invalid scope format
feat(-T1): Invalid scope format          # Scope can't start with hyphen
feat(T1-): Invalid scope format          # Scope can't end with hyphen
feat(app//api): Invalid scope format     # Scope can't contain empty slash segments
feat(scope): not capitalized             # Invalid with --description-rules=capitalized
feat(scope): Summary with 60+ chars...   # Invalid with --description-length-limit=60
```

## Contributing

Feel free to open issues and pull requests!
