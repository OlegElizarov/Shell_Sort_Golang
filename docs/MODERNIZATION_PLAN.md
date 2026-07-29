# Modernization Plan: Shell_Sort_Golang

Status: not yet executed. Written 2026-07-14, to be implemented later.

## Context

Repo is a small Go library (module `github.com/OlegElizarov/Shell_Sort_Golang`, `go 1.26`) implementing Shell Sort with goroutine-parallelized subarray sorting. Current state, confirmed via `gopls` navigation:

- `lib/sort.go` — `ShellSort([]int) []int`, `selectStepSedgewick`, `selectStepHibbard` (dead — only called from its own test), `subSort`. Contains a TODO (`const for choosing select func`), commented-out bubble sort, commented-out trim logic, commented-out duplicate swap.
- `lib/slices.go` — hardcoded fixture slices (16/100/500 elements) used by both tests and `main.go`.
- `lib/sort_test.go` — table-driven tests against hardcoded `*Correct` literals; benchmarks (`BenchmarkshellSort`/`BenchmarkstdSort`) reuse a **single shared slice mutated by concurrent `b.RunParallel` workers with no synchronization** — a real data race and a measurement bug (after first sort, subsequent iterations measure already-sorted input). Only 2 of 8 declared benchmarks are active; one duplicate function name among the dead stubs.
- `main.go` — demo entrypoint at repo root, package `lib` sits in a subdirectory (discouraged generic name).

Goal: (1) modernize/refactor code to current Go idioms, (2) redesign tests/benchmarks to drop hardcoded slices while still comparing ShellSort against stdlib on **identical** generated input, (3) restructure folders per Go library conventions. Decisions confirmed with user: expose configurability as **public functional options**, move package to **repo root** with demo in `examples/`, keep Hibbard as a **selectable alternative** rather than deleting it.

## A. Code refactor (`lib/sort.go` → root `sort.go`)

- **Generic API**, matching stdlib's own shape:
  ```go
  func ShellSort[S ~[]E, E cmp.Ordered](x S, opts ...Option) S
  func subSort[T cmp.Ordered](x []T, gap, start int) // no pointer, no *sync.WaitGroup param
  ```
  Type inference keeps existing int-slice call sites compiling unchanged.
- **Functional options**, resolving the TODO at `sort.go:9` and making gap-sequence/parallelism configurable for both library consumers and benchmarks:
  ```go
  type Option func(*config)
  func WithGapSequence(f GapSequenceFunc) Option
  func WithParallelThreshold(n int) Option // default 5

  type GapSequenceFunc func(length int) []int
  func SedgewickGaps(length int) []int // exported, was selectStepSedgewick, default
  func HibbardGaps(length int) []int   // exported, was selectStepHibbard, opt-in
  ```
- **Dead code removed**: commented bubble sort, commented trim lines, commented duplicate swap. Gap-sequence functions rewritten to `append`-only (no `make([]int, 10)` pre-sizing) — eliminates the latent zero-padding footgun that currently makes short gap sequences work only by accident.
- **Parallel/sequential branches consolidated** into one loop gated by the configurable threshold, instead of two near-duplicate loops differing only by `go`/`wg`.
- **Robustness**: explicit `if len(x) < 2 { return x }` early return; doc comment states in-place mutation explicitly (mirrors `slices.Sort` style).
- Every new exported identifier (`Option`, `WithGapSequence`, `WithParallelThreshold`, `GapSequenceFunc`, `SedgewickGaps`, `HibbardGaps`) gets a doc comment — keeps the current zero-lint-diagnostic baseline.

## B. Tests & benchmarks (main ask)

New files at repo root (package `shellsort`): `sort_test.go`, `sort_bench_test.go`, `fuzz_test.go`. `lib/slices.go` deleted — no hardcoded fixtures survive.

- **Shared seeded generator**, extending the pattern already in the repo (`rand.NewPCG` is already used for benchmarks — just centralized and reused everywhere instead of re-invented):
  ```go
  func randomSlice(tb testing.TB, n int, seed1, seed2 uint64) []int {
      tb.Helper()
      return rand.New(rand.NewPCG(seed1, seed2)).Perm(n)
  }
  ```
- **`TestShellSort`**: table of `{Name, Size}` (empty, single, short=16, medium=100, large=500, plus explicit already-sorted and reverse-sorted cases). Correctness oracle is `slices.Sort` on a clone, not a hardcoded literal — `want := slices.Clone(input); slices.Sort(want)` vs `got := ShellSort(slices.Clone(input))`. Explicit cloning fixes a real bug in the current test where `ShellSort`'s in-place mutation makes the second assertion (`slices.IsSorted(ShellSort(tc.Input))`) trivially pass on already-sorted data.
- **`TestSedgewickGaps`/`TestHibbardGaps`**: replace non-reproducible `rand.IntN` calls with a fixed table of boundary lengths (0, 1, 2, 16, large) instead of one random point per run.
- **`BenchmarkSort`** in `sort_bench_test.go`: one `base := randomSlice(b, n, 1, 2)` per size in `benchSizes := []int{1_000, 10_000, 20_000, 30_000, 40_000, 50_000}`, then pre-cloned slices per `b.N` before `b.ResetTimer()`, with sub-benchmarks:
  - `ShellSort/n=%d` vs `StdSort/n=%d` — same `base` per size, satisfying "compare on identical input."
  - `ShellSort/parallel/n=%d` vs `ShellSort/sequential/n=%d` (via `WithParallelThreshold(0)`) — makes the parallelism-worth-it question answerable directly from `go test -bench` output.
  - `ShellSort/sedgewick/n=%d` vs `ShellSort/hibbard/n=%d` (via `WithGapSequence(HibbardGaps)`) — payoff of keeping Hibbard as selectable.
  - This replaces all 8 old benchmark functions (fixes the data-race/measurement bug and the duplicate-name bug) with one parametrized suite.
- **`FuzzShellSort`** in `fuzz_test.go` (native `testing.F`, stdlib, no new dependency): fuzzes `(seed int64, n uint8)`, generates a bounded-range slice (covers duplicates/ties, which `Perm` structurally can't), asserts against `slices.Sort` oracle. Seed corpus includes empty and single-element cases. Runs automatically under plain `go test ./...` — free regression coverage in CI, no workflow changes needed.

## C. Folder structure

- Delete `lib/slices.go`.
- Move+rename `lib/sort.go` → `sort.go` (repo root), `package lib` → `package shellsort`, apply all Area A changes.
- Move+split `lib/sort_test.go` → `sort_test.go` + `sort_bench_test.go` + `fuzz_test.go` (repo root, `package shellsort`).
- Delete now-empty `lib/` directory.
- Move+rename `main.go` → `examples/shellsort/main.go`; `package main`; import path becomes `github.com/OlegElizarov/Shell_Sort_Golang` (aliased `shellsort` since the last path segment won't lexically match — normal in Go); demo builds its own small seeded slice inline instead of importing removed fixtures.
- Update `README.md` (new import path, `go run ./examples/shellsort`) and `CLAUDE.md` (paths, package name, resolved TODO/Hibbard status, new commands).
- No changes needed to `.github/workflows/ci.yml` or `go.mod` — CI already globs `./...`, module path unchanged.

## Critical files
- `lib/sort.go` → `sort.go`
- `lib/sort_test.go` → `sort_test.go` / `sort_bench_test.go` / `fuzz_test.go`
- `lib/slices.go` (deleted)
- `main.go` → `examples/shellsort/main.go`
- `README.md`, `CLAUDE.md`

## Verification (when executed)
- `go build ./...`, `go vet ./...` — clean build across new layout.
- `go test -v -race ./...` — table tests, fuzz seed corpus, and no benchmark data race.
- `go test -bench=. -benchmem ./...` — confirm ShellSort vs StdSort, parallel vs sequential, and Sedgewick vs Hibbard sub-benchmarks all run across the full size matrix.
- `golangci-lint run --timeout=5m` — zero new diagnostics.
- `gopls check ./...` — zero diagnostics, confirming doc-comment coverage on all new exported identifiers.
- `go run ./examples/shellsort` — demo still runs and prints sorted output.
