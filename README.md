# Commit Message Guardian

A Git pre-commit hook that validates commit messages against specified rules, ensuring consistent message formatting and language usage.

## Features

- Validates commit messages against the [Conventional Commits](https://www.conventionalcommits.org/) format
- Supports various text validation rules:
  - `noCyrillic`: Prevents Cyrillic characters
  - `noLatin`: Prevents Latin characters
  - `noDigits`: Prevents digits
  - `cyrillicOnly`: Allows only Cyrillic characters
  - `latinOnly`: Allows only Latin characters
  - `digitsOnly`: Allows only digits
- Configurable validation rules for different parts of the commit message (type, scope, description)
- Body text is not validated by default

## Installation

To use this hook in your project:

1. Install [pre-commit](https://pre-commit.com/) if you haven't already
2. Add this to your `.pre-commit-config.yaml`:

```yaml
repos:
- repo: github.com/AnruKitakaze/commit-msg-guardian
  rev: v1.0.0  # Use the latest version
  hooks:
    - id: commit-msg-guardian
      # Optional: override default rules
      args:
        - --type-rules=latinOnly
        - --scope-rules=latinOnly,digitsOnly
        - --description-rules=noCyrillic
```

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

- `--type-rules`: Comma-separated rules for commit type (default: "latinOnly")
- `--scope-rules`: Comma-separated rules for commit scope (default: "latinOnly,digitsOnly")
- `--description-rules`: Comma-separated rules for commit description (default: "noCyrillic")

### Examples

Valid commit messages:
```
feat(TGK-1827): This is an example
docs(T-1): This is valid too

With кириллица in description (body is not validated by default)
```

Invalid commit messages:
```
feat(TGK-1827): Забыл убрать кириллицу   # Contains Cyrillic in description
random: Not a valid type                 # Invalid commit type
feat[scope]: Wrong scope format          # Invalid scope format
```

## Contributing

Feel free to open issues and pull requests!
