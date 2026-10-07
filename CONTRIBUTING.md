# Contributing to Coctyl

Thank you for your interest in contributing to Coctyl! Coctyl is a structural fingerprinting CLI and library for Go source code ("dactyloscopy") that hashes functional logic while ignoring superficial labels and formatting.

By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).

---

## Table of Contents

- [How Can I Contribute?](#how-can-i-contribute)
  - [Reporting Bugs](#reporting-bugs)
  - [Suggesting Enhancements](#suggesting-enhancements)
  - [Submitting Pull Requests](#submitting-pull-requests)
- [Development Environment Setup](#development-environment-setup)
  - [Prerequisites](#prerequisites)
  - [Initial Setup](#initial-setup)
- [Development Workflow](#development-workflow)
  - [Building](#building)
  - [Testing](#testing)
  - [Linting & Formatting](#linting--formatting)
- [Architecture & Two-Pass Pipeline](#architecture--two-pass-pipeline)
- [Pull Request Guidelines](#pull-request-guidelines)
- [License Notice](#license-notice)

---

## How Can I Contribute?

### Reporting Bugs

Before reporting an issue, please search existing GitHub issues to verify it hasn't already been addressed.

If you find a new bug (e.g. false equivalence or unintended divergence in dactyls), please file an issue including:
- A clear, descriptive title.
- Minimal Go source snippets demonstrating the unexpected behavior.
- Expected hash behavior vs. actual hash output.
- Environment details (OS, Go version).

### Suggesting Enhancements

We welcome suggestions for refining AST normalization, performance improvements, and CLI capabilities!
- Check existing issues and discussions before opening a proposal.
- Clearly describe the use case and rationale.

### Submitting Pull Requests

Small fixes, documentation improvements, and additional test cases are welcome directly as Pull Requests. For major features or AST normalization changes, opening an issue first is recommended.

---

## Development Environment Setup

### Prerequisites

- [Go](https://go.dev) 1.27 or newer
- `make` (optional, for running Makefile targets)

### Initial Setup

1. Fork and clone the repository:
   ```bash
   git clone https://github.com/<your-username>/coctyl.git
   cd coctyl
   ```
2. Verify tests:
   ```bash
   go test ./...
   ```

---

## Development Workflow

### Building

Build the `coctyl` CLI binary:
```bash
make build
# or directly:
go build -o bin/coctyl ./cmd/coctyl
```

### Testing

Always run tests before submitting a pull request:
```bash
# Run all unit, integration, and validation tests
make test

# Generate statement test coverage report
make test-coverage
```

### Linting & Formatting

Coctyl enforces standard Go formatting and vetting:
```bash
# Auto-format Go source files
make fmt

# Run Go static vetting
make vet
```

---

## Architecture & Two-Pass Pipeline

Coctyl computes structural dactyls using a clean two-pass pipeline:

1. **Pass 1 — AST Normalization (`pkg/coctyl/normalize.go`):**
   - Resolves lexical scopes and alpha-normalizes local variables and parameters to scope-order indices (`$0, $1, ...`).
   - Distinguishes predeclared builtins (`builtin:len`) from external identifiers (`ident:Foo`).
   - Ignores package headers and imports by default for functional equivalence.
   - Strips comments and AST wrappers (`*ast.ParenExpr`).

2. **Pass 2 — Stateless Serialization (`pkg/coctyl/serialize.go`):**
   - Stateless, immutable printer converting the normalized AST directly into a canonical S-expression string.

3. **Hashing (`pkg/coctyl/hasher.go`):**
   - Computes deterministic SHA-256 dactyl from the canonical S-expression string.

---

## Pull Request Guidelines

1. **Branch Naming**: Use descriptive branch names like `feat/support-labels`, `fix/closure-counter`, or `docs/readme-diagram`.
2. **Atomic Commits**: We recommend [Conventional Commits](https://www.conventionalcommits.org/):
   - `feat: add flag for package inclusion`
   - `fix: resolve composite literal key binding`
   - `test: add equivalence test for naked returns`
3. **Keep CI Green**: Ensure `make vet` and `make test` pass cleanly before submitting.
4. **Validation Tests**: For any change to AST normalization, add counterpart tests in `pkg/coctyl/validation_tests/`.

---

## License Notice

Coctyl is licensed under the [MIT License](LICENSE). By contributing to Coctyl, you agree that your contributions will be licensed under the terms of the MIT License.