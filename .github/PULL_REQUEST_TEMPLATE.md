## Related Issue

Closes #<!-- issue number -->

> **Note**: PRs without a related Issue may be closed. Please [open an Issue](https://github.com/gpdf-dev/gpdf/issues/new/choose) first to discuss your approach.

## Changes

<!-- What does this PR do? -->

## Checklist

- [ ] Related Issue is linked above
- [ ] `go test ./...` passes
- [ ] `go test -race ./...` passes
- [ ] `go vet ./...` passes
- [ ] No external dependencies added
- [ ] New public APIs have GoDoc comments
- [ ] Benchmarks are not degraded (`cd _benchmark && go test -bench=. -benchmem`)
