# Contributing to gpdf

Thank you for your interest in gpdf! This document explains how to contribute.

## How to Contribute

### Reporting Bugs & Requesting Features

The best way to contribute is by **opening an Issue**:

- **Bug reports** — describe the problem, expected behavior, and steps to reproduce
- **Feature requests** — describe the use case and why it would be useful
- **Questions** — feel free to ask in Issues

### Pull Requests

> **Important**: Please **open an Issue first** before submitting a Pull Request.

gpdf has strict design principles that are easy to violate unintentionally. To avoid wasted effort, discuss your approach in an Issue before writing code. PRs without prior discussion may be closed.

Small fixes (typos, documentation improvements) can be submitted directly.

## Design Principles

All contributions must respect these principles:

1. **Zero dependencies** — gpdf uses only the Go standard library. Do not add external packages.
2. **Layered architecture** — `template` → `document` → `pdf` (one-way only). Layer 3 must not import Layer 1.
3. **Performance** — gpdf is 10-30x faster than alternatives. PRs must not degrade benchmark results.
4. **CJK first-class** — Japanese, Chinese, and Korean text must always be considered.

## Development Setup

```bash
git clone https://github.com/gpdf-dev/gpdf.git
cd gpdf
go test ./...
```

### Running Tests

```bash
go test ./...              # All tests
go test -race ./...        # Race detector
go vet ./...               # Static analysis
```

### Running Benchmarks

```bash
cd _benchmark && go test -bench=. -benchmem -count=5
```

## Code Style

- Follow standard `gofmt` formatting (enforced by CI)
- All public APIs must have GoDoc comments
- Cyclomatic complexity must not exceed 15 (enforced by CI)
- Test files use `_test.go` suffix
- Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/): `feat:`, `fix:`, `refactor:`, `test:`, `docs:`, `chore:`

## PR Checklist

Before submitting a PR:

- [ ] Related Issue is linked
- [ ] `go test ./...` passes
- [ ] `go test -race ./...` passes
- [ ] `go vet ./...` passes
- [ ] No external dependencies added
- [ ] New public APIs have GoDoc comments
- [ ] Benchmarks are not degraded

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
