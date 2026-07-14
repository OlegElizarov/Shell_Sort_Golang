# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

- Build: `go build -v ./...`
- Vet: `go vet ./...`
- Test all (matches CI): `go test -v -race -coverprofile=coverage.out ./...`
- Run a single test: `go test -run TestShellSort ./lib/...`
- Run a single subtest: `go test -run 'TestShellSort/Short_slice_test' ./lib/...`
- Lint: `golangci-lint run --timeout=5m` (uses `.golangci.yml`, v2 schema — requires golangci-lint v2+; enables errcheck, govet, ineffassign, staticcheck, unused, gocritic, paralleltest, revive plus gofmt formatting)
- Run demo: `go run .`
- Benchmark + profile: `go test -bench=. -benchmem -memprofile=mem -cpuprofile=cpu .`, then `go tool pprof -alloc_objects -http=:8080 mem` (or `go tool pprof -alloc_objects mem` for terminal)

CI (`.github/workflows/ci.yml`) runs a build/vet/test matrix across Go 1.23/1.24/1.26 on ubuntu/macos/windows, plus a separate `golangci-lint` job.

## Architecture

- `main.go` — plain demo entrypoint (not a CLI); sets `GOMAXPROCS`, prints `lib.ShellSort` output on the fixture slices.
- `lib/sort.go` — the only exported function is `ShellSort([]int) []int`. It splits the slice by gap sequence and insertion-sorts each subarray, in parallel via goroutines (`sync.WaitGroup`) when there are fewer than 5 subarrays, sequentially otherwise (parallelizing 5+ subarrays doesn't pay off given typical core counts).
- Gap sequence is chosen at compile time, not via a runtime flag: `selectStepSedgewick` is active; `selectStepHibbard` is a working alternative left commented out in `ShellSort`. Swapping requires editing the source.
- `lib/slices.go` — shared exported fixture vars (`IntSliceShort`, `IntSliceBig`, `IntSliceVeryBig`, and their `*Correct` sorted counterparts) used by both `lib/sort_test.go` and `main.go`. Reuse these instead of inventing new literal slices when adding tests.
- `lib/sort_test.go` — table-driven tests using `testify/require`, with `t.Parallel()` at both the top-level test and each subtest. Benchmark helpers use a non-standard signature (`func(i int, b *testing.B)`) wrapped by real `Benchmark*(b *testing.B)` entry points; most size variants are commented out except the active 30000-element ones.
