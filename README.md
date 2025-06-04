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
- Configurable validation rules for different parts of the commit message (type, scope, description)
- Body text is not validated by default

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
        - --description-rules=noCyrillic
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

### Examples

Valid commit messages:
```
feat(TGK-1827): This is an example
docs(T-1): This is valid too
feat(T1): Another valid example
feat(task-123): Valid with hyphen

With кириллица in description (body is not validated by default)
```

Invalid commit messages:
```
feat(TGK-1827): Забыл убрать кириллицу   # Contains Cyrillic in description
random: Not a valid type                 # Invalid commit type
feat[scope]: Wrong scope format          # Invalid scope format
feat(-T1): Invalid scope format          # Scope can't start with hyphen
feat(T1-): Invalid scope format          # Scope can't end with hyphen
```

## Contributing

Feel free to open issues and pull requests!
