# Modernization Plan: Shell_Sort_Golang

Status: partially executed. Written 2026-07-14, revised 2026-08-03 to fold in the findings of [GAP_SEQUENCES.md](GAP_SEQUENCES.md).

## Context

Repo is a small Go library (module `github.com/OlegElizarov/Shell_Sort_Golang`, `go 1.26`) implementing Shell Sort with goroutine-parallelized subarray sorting.

**Already done since the first draft of this plan:**

- Demo moved out of the repo root to `cmd/main.go`.
- `lib/sort_test.go` benchmarks rebuilt: seeded `rand.NewPCG` generator, one shared base permutation per size (`benchInput`), fresh clone per iteration, `b.Loop`, `b.ReportAllocs`, pinned `GOMAXPROCS`. The old data race and already-sorted-input measurement bug are gone, as are the duplicate/dead benchmark stubs.
- `TestSelectStepSedgewick` / `TestSelectStepHibbard` exist.

**Still outstanding:**

- `lib/sort.go` — `ShellSort([]int) []int` is `int`-only and non-configurable; gap sequence is chosen by editing line 15. `selectStepHibbard` is dead outside its own test. The `TODO: const for choosing select func` at line 9 is unresolved. Commented-out bubble sort and trim logic remain.
- Both gap functions pre-size with `make([]int, 10)` and index into it, so a sequence shorter than 10 entries returns trailing zeros. It works today only because a zero gap in `subSort` is never reached with the current inputs — a latent footgun, not a design.
- `lib/slices.go` — hardcoded fixtures still used by `TestShellSort` and `cmd/main.go`.
- Package is still `lib` in a subdirectory (discouraged generic name).
- `TestShellSort` asserts against hardcoded `*Correct` literals and calls `ShellSort(tc.Input)` twice on the same slice — the second assertion is trivially true because the first call sorted in place.

Goals: (1) modernize to current Go idioms, (2) make the gap sequence a first-class runtime choice with a catalog of implementations, (3) build a benchmark study that can actually answer which sequence is best *for this implementation*, (4) restructure per Go library conventions.

Decisions confirmed with user: expose configurability as **public functional options**, move package to **repo root** with demo in `examples/`, keep Hibbard as a **selectable alternative** rather than deleting it.

## A. Core API (`lib/sort.go` → root `sort.go`)

- **Generic API**, matching stdlib's shape:
  ```go
  func ShellSort[S ~[]E, E cmp.Ordered](x S, opts ...Option) S
  func subSort[T cmp.Ordered](x []T, gap, start int) // no pointer, no *sync.WaitGroup param
  ```
  Type inference keeps existing int-slice call sites compiling unchanged.
- **Functional options**, resolving the TODO at `sort.go:9`:
  ```go
  type Option func(*config)
  func WithGapSequence(g GapSequence) Option
  func WithParallelThreshold(minSubarrayLen int) Option // see section E
  ```
- **Gap sequence as a named value**, not a bare func — the benchmark suite (section D) needs to label results, and several sequences carry provenance worth surfacing:
  ```go
  type GapSequence struct {
      Name string
      Gaps func(length int) []int // ascending, gaps[0] == 1, all < length
  }
  ```
  The `length` parameter is required, not incidental: Shell, Frank–Lazarus and Gonnet–Baeza-Yates derive their gaps *from* `N` rather than generating bottom-up, and Ciura's fixed tables need `N` to decide where to start extending.
- **Contract for every implementation**, enforced by the shared test in section C: ascending order, `gaps[0] == 1`, strictly increasing, every element `< length`, never empty (a `length < 2` input still yields `[]int{1}`).
- **Dead code removed**: commented bubble sort, commented trim lines, commented duplicate swap. All generators rewritten `append`-only — no `make([]int, 10)` pre-sizing, killing the zero-padding footgun.
- Every exported identifier gets a doc comment, preserving the current zero-lint baseline.

## B. Gap sequence catalog

Implement all of the following in `gaps.go`. Each is a handful of lines; the point of having them all is that section D's study is only meaningful across the full field, including the known-bad ones as controls. Analysis and citations for each: [GAP_SEQUENCES.md](GAP_SEQUENCES.md).

| Exported name | Formula | First terms | Role |
| --- | --- | --- | --- |
| `Tokuda` | `⌈h'⌉`, `h' = 2.25h' + 1`, `h'_1 = 1` | 1, 4, 9, 20, 46, 103, 233, 525 | **default** |
| `Ciura` | fixed table + `⌊2.25h⌋` extension | 1, 4, 10, 23, 57, 132, 301, 701, 1750 | best measured comparisons |
| `Ciura128` | fixed table, searched for `N = 128` | 1, 4, 9, 24, 85, 126 | small-`N` reference |
| `Ciura1000` | fixed table, searched for `N = 1000` | 1, 4, 10, 23, 57, 156, 409, 995 | mid-`N` reference |
| `Knuth` | `3h + 1`, capped at `⌈N/3⌉` | 1, 4, 13, 40, 121, 364 | integer-only fallback |
| `Lee` | `⌈(γ^k − 1)/(γ − 1)⌉`, γ = 2.243609061420001 | 1, 4, 9, 20, 45, 102, 230, 516 | tuned Tokuda variant |
| `SkeanB` | `⌊a·b^(i/c) + d⌋`, a=4.0816, b=8.5714, c=2.2449, d=0 | 1, 4, 10, 27, 72, 187, 488 | ratio 2.604, large-`N` comparisons |
| `Pratt` | 3-smooth `2^p·3^q` | 1, 2, 3, 4, 6, 8, 9, 12, 16 | best worst case; **best exchanges** |
| `Pratt25` | `2^p·5^q` | 1, 2, 4, 5, 8, 10, 15, 16 | best exchanges measured |
| `Pratt34` | `3^p·4^q` | 1, 3, 4, 9, 12, 16, 24 | Pratt family, fewer comparisons |
| `Sedgewick86` | piecewise even/odd `k` | 1, 5, 19, 41, 109 | current default, keep for continuity |
| `Sedgewick82` | `4^k + 3·2^{k−1} + 1` | 1, 8, 23, 77, 281 | ratio-4 datapoint |
| `Hibbard` | `2^k − 1` | 1, 3, 7, 15, 31 | keep per user decision |
| `PapernovStasevich` | `2^k + 1`, prefixed 1 | 1, 3, 5, 9, 17, 33 | ratio-2 control |
| `IncerpiSedgewick` | constructed, coprime triangle | 1, 3, 7, 21, 48, 112 | best asymptotic exponent |
| `FrankLazarus` | `2⌊N/2^{k+1}⌋ + 1` | …, 3, 1 (`N`-derived) | `N`-dependent control |
| `GonnetBaezaYates` | `max(⌊(5h−1)/11⌋, 1)` from `h_0 = N` | 1, 3, 8, 19, 42, 93, 206, 454 (`N`=1000) | ratio-2.2, `Θ(N²)` for some `N` |
| `Shell` | `⌊N/2^k⌋` | N/2, N/4, …, 1 | negative control (`Θ(N²)`) |

Notes for implementation:

- `Pratt*` families are generated by merging two geometric progressions and sorting, not by a recurrence. Generating count is `Θ(log²N)` — the only family whose gap slice is large enough to be worth pre-sizing.
- `Ciura*` need a documented extension rule past their last tabulated term; `⌊2.25h⌋` is the convention adopted by Skean et al., and is **unvalidated** — say so in the doc comment.
- `Lee` and `SkeanB` need `math.Pow`; `Knuth`, `Hibbard`, `PapernovStasevich`, `Pratt*` and `GonnetBaezaYates` are integer-only. Worth noting in doc comments for anyone avoiding float.
- `Shell` and `GonnetBaezaYates` are the two that can produce genuinely bad sequences. Keep them; the study needs a floor.

## C. Tests

New files at repo root (package `shellsort`): `sort_test.go`, `gaps_test.go`, `sort_bench_test.go`, `fuzz_test.go`. `lib/slices.go` deleted.

- **Shared seeded generator**, replacing the fixtures:
  ```go
  func randomSlice(tb testing.TB, n int, seed1, seed2 uint64) []int {
      tb.Helper()
      return rand.New(rand.NewPCG(seed1, seed2)).Perm(n)
  }
  ```
- **`TestShellSort`**: table of `{Name, Size}` (empty, single, 16, 100, 500, plus explicit already-sorted and reverse-sorted). Oracle is `slices.Sort` on a clone, not a literal: `want := slices.Clone(in); slices.Sort(want)` vs `got := ShellSort(slices.Clone(in))`. The explicit cloning fixes the in-place-mutation bug in the current test.
- **`TestShellSortAllSequences`**: run the correctness table against **every** sequence in the catalog. This is the payoff of section A's `GapSequence` contract — one loop covers 18 implementations.
- **`TestGapSequenceContract`**: for each catalog entry and each `N` in `{0, 1, 2, 16, 128, 1000, 50000}`, assert ascending, strictly increasing, `gaps[0] == 1`, all `< N` (for `N > 1`), non-empty. Replaces the current per-function tests and their non-reproducible `rand.IntN` lengths.
- **`TestGapSequenceGoldenTerms`**: assert the exact first terms in section B's table. These are transcribed from primary sources and are the only defense against a plausible-looking formula that silently generates the wrong sequence — the failure mode that cost the most time while writing GAP_SEQUENCES.md.
- **`FuzzShellSort`** (`testing.F`, stdlib): fuzz `(seed int64, n uint8)`, bounded value range so duplicates and ties appear (which `Perm` structurally cannot produce), assert against the `slices.Sort` oracle. Seed corpus includes empty and single-element. Runs under plain `go test ./...`.

## D. Benchmark study

The central question — which sequence is fastest *here* — is not answerable from the literature. Published rankings are dominated by comparison counts measured in Python, and the top group is separated by under 2% (GAP_SEQUENCES.md §5.1), well inside what goroutine overhead and cache behaviour will swamp in this implementation.

**Three metrics, measured separately.** They disagree, and that disagreement is the finding:

1. **Comparisons** — literature-comparable, implementation-independent.
2. **Exchanges** — Pratt beats Tokuda by ~30% here while losing 3.1× on comparisons.
3. **Wall-clock + allocs/op** — the only metric that decides the default.

**Counting harness.** Comparisons and exchanges need instrumentation that must not exist on the production path. Add an unexported `sortInstrumented` in `gaps_bench_test.go` (test files are excluded from the shipped package) that mirrors `subSort` with two counters. Match Skean et al.'s definitions exactly or the numbers are not comparable: a **comparison** is one evaluation of `A(i) > A(i+k)`; an **exchange** is one swap performed to fix an inversion. The current `subSort` uses a shifting insertion sort, so decide explicitly whether a shift counts as an exchange and record the choice in a comment.

**Harness validation — do this first.** Run the counting harness on Tokuda, Ciura, Pratt-23 and Pratt-25 at `N = 10 000` and compare against the published values:

| Sequence | Expected μCO | Expected μEX |
| --- | --- | --- |
| Tokuda | 192 574 | 98 071 |
| Ciura (Large) | 191 435 | 101 680 |
| Pratt-23 | 604 502 | 66 923 |
| Pratt-25 | 450 131 | 62 191 |

These counts are implementation-independent given matching definitions, so agreement within a percent or two validates both the gap generators and the counter. **Disagreement means a bug in the catalog, not a discovery** — this is the cheapest correctness check available for the whole of section B, and it should gate everything downstream.

**Benchmark matrix** in `sort_bench_test.go`, extending the existing `benchInput`/`benchmarkSort` helpers:

- `Sort/{sequence}/n={size}` across the full catalog, `benchSizes = {1_000, 10_000, 100_000}`.
- `Sort/std/n={size}` — `slices.Sort` baseline on the identical base permutation.
- `Sort/{parallel,sequential}/{best sequence}/n={size}` — quantifies the goroutine machinery against `WithParallelThreshold` disabled. GAP_SEQUENCES.md §6 predicts a modest ceiling here; this measures it.
- Input shapes beyond random permutations: already-sorted, reverse-sorted, few-unique. Weiss's `O(N log N)` result for reverse-ordered input means the adversarial case for plain insertion sort is *not* adversarial for Shellsort, and the suite should show that.

**Measurement hygiene:** pre-clone all `b.N` inputs before `b.ResetTimer()` rather than calling `b.StopTimer()`/`b.StartTimer()` inside the loop — the current helper pays timer-toggle overhead on every iteration, which is significant at these sizes. Report with `benchstat` over `-count=10`.

## E. Parallel redesign

Two constructs in the current code are unjustified, and GAP_SEQUENCES.md §6 explains why:

- **`if d[...] < 5` (`lib/sort.go:19`)** — parallelizes only when there are *fewer* than 5 subarrays, i.e. exactly when there is least parallelism available and each subarray is longest. The comment cites core count, but the effect is that the wide early passes run sequentially. Replace with a threshold on **work per subarray** (`len(x)/gap >= minSubarrayLen`), so parallelism follows the work rather than the subarray count, and expose it as `WithParallelThreshold`.
- **`d*3 < float64(length)` (`lib/sort.go:41`)** — an undocumented cap that silently drops the largest gaps. Knuth's own rule is to stop at `⌈N/3⌉`, which is a *stopping* condition for the generator, not a filter applied to every element. Fold it into the `Knuth` generator where it belongs and drop it from the shared path.

Additionally, **decouple goroutine count from gap count**. Spawning one goroutine per subarray means the first pass of a closed-form sequence spawns thousands of goroutines that each sort a handful of elements. Chunk instead: partition the `gap` subarrays across `min(gap, GOMAXPROCS)` workers. This is independent of sequence choice and is likely the single largest available win — worth benchmarking before and after, as its own line in the study.

Note what will *not* improve: the final `h = 1` pass is a single subarray of `N` elements and is unparallelizable, so span is dominated by the low-width tail passes regardless of sequence or chunking. Amdahl caps the whole approach. Sequence choice changes barrier count (~17 for Tokuda vs ~125 for Pratt at `N = 10^6`), not the floor.

## F. Folder structure

- Move+rename `lib/sort.go` → `sort.go` (repo root), `package lib` → `package shellsort`, apply section A.
- New `gaps.go` (catalog) + `gaps_test.go`.
- Move+split `lib/sort_test.go` → `sort_test.go` + `sort_bench_test.go` + `fuzz_test.go`.
- Delete `lib/slices.go` and the now-empty `lib/`.
- Move `cmd/main.go` → `examples/shellsort/main.go`; import path becomes `github.com/OlegElizarov/Shell_Sort_Golang` (aliased `shellsort`, since the last path segment won't lexically match — normal in Go). Demo builds a seeded slice inline instead of importing removed fixtures.
- Update `README.md` (import path, `go run ./examples/shellsort`, sequence catalog) and `CLAUDE.md` (paths, package name, resolved TODO, new commands). Note that `CLAUDE.md` currently claims `main.go` is at the repo root — already stale.
- No changes needed to `.github/workflows/ci.yml` or `go.mod`: CI globs `./...`, module path unchanged.

## G. Suggested execution order

Each step leaves the tree building and testing green.

1. **Catalog first, in place.** Add `GapSequence` + the 18 generators to the existing `lib` package, with the contract test and golden-terms test. No API change yet. Fastest path to a testable artifact, and flushes out formula bugs while the sources are fresh.
2. **Counting harness + literature validation** (section D). Gate: the four published values reproduce. Do not proceed past a mismatch.
3. **Options API** — `WithGapSequence`, `WithParallelThreshold`, generics, dead-code removal, `append`-only generators.
4. **Test rebuild** — oracle-based `TestShellSort`, all-sequences correctness loop, fuzz target.
5. **Benchmark matrix** and `benchstat` run. Record results in a new `docs/BENCHMARKS.md`; the empirical answer belongs next to the literature survey.
6. **Parallel redesign** (section E) — re-run the matrix, before/after, as a separate commit so the effect is attributable.
7. **Restructure** (section F) — pure moves, last, so no behavioural change hides inside a rename.
8. **Docs** — update `README.md`, `CLAUDE.md`, and add measured results to GAP_SEQUENCES.md §5.1 alongside the published ones.

Steps 1–2 are worth doing even if the rest is deferred: they turn the literature survey into executable, verified code.

## Critical files

- `lib/sort.go` → `sort.go`, new `gaps.go`
- `lib/sort_test.go` → `sort_test.go` / `gaps_test.go` / `sort_bench_test.go` / `fuzz_test.go`
- `lib/slices.go` (deleted)
- `cmd/main.go` → `examples/shellsort/main.go`
- `README.md`, `CLAUDE.md`, `docs/GAP_SEQUENCES.md`, new `docs/BENCHMARKS.md`

## Verification

- `go build ./...`, `go vet ./...` — clean across the new layout.
- `go test -v -race ./...` — table tests, all-sequences loop, contract + golden terms, fuzz seed corpus, no benchmark data race.
- `go test -bench=. -benchmem ./...` — full matrix runs across sizes, sequences, input shapes, and parallel/sequential.
- **Literature reproduction** (section D) — the gate on the whole catalog.
- `golangci-lint run --timeout=5m` — zero new diagnostics.
- `gopls check ./...` — zero diagnostics, confirming doc-comment coverage on ~20 new exported identifiers.
- `go run ./examples/shellsort` — demo still prints sorted output.
