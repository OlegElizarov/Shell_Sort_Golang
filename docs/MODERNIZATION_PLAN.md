# Modernization Plan: Shell_Sort_Golang

Written 2026-07-14. Revised 2026-08-03 to fold in [GAP_SEQUENCES.md](GAP_SEQUENCES.md) and to re-cut the work as gated, individually shippable phases.

---

## Progress

`▰▰▰▰▱▱▱▱▱▱` **4 / 10 phases complete** — 28 / 66 tasks

| # | Phase | Tasks | Status | Gate |
| --- | --- | --- | --- | --- |
| 0 | [Baseline & prep](#phase-0--baseline--prep) | 5 | ✅ done | tree green |
| 1 | [Gap catalog, in place](#phase-1--gap-catalog-in-place) | 8 | ✅ done | catalog compiles |
| 2 | [Contract + golden-terms tests](#phase-2--contract--golden-terms-tests) | 8 | ✅ done | 18 sequences pass contract |
| 3 | [Counting harness + literature validation](#phase-3--counting-harness--literature-validation) | 7 | ✅ done | **4 published counts reproduce** |
| 3.1 | [Parked nits and small fixes](#phase-31--parked-nits-and-small-fixes) | 2 | 🅿️ backlog | absorbed by later phases |
| 4 | [Options API + generics](#phase-4--options-api--generics) | 8 | ⬜ not started | old call sites still compile |
| 5 | [Test rebuild](#phase-5--test-rebuild) | 6 | ⬜ not started | oracle-based, race-clean |
| 6 | [Benchmark matrix](#phase-6--benchmark-matrix) | 8 | ⬜ not started | `docs/BENCHMARKS.md` exists |
| 7 | [Parallel redesign](#phase-7--parallel-redesign) | 6 | ⬜ not started | before/after benchstat |
| 8 | [Restructure to root package](#phase-8--restructure-to-root-package) | 6 | ⬜ not started | pure moves, tests unchanged |
| 9 | [Docs](#phase-9--docs) | 4 | ⬜ not started | `gopls check` clean |

Status legend: ⬜ not started · 🟨 in progress · ✅ done · ⛔ blocked · 🅿️ backlog (not gating; drained by the phase that names it)

**Updating this file is part of each phase.** Tick the boxes as you go, flip the row's status, and redraw the bar (`▰` per complete phase). One commit per phase, message given in the phase header.

---

## Context

Small Go library (module `github.com/OlegElizarov/Shell_Sort_Golang`, `go 1.26`) implementing Shell Sort with goroutine-parallelized subarray sorting.

**Done since the first draft:**

- Demo moved out of repo root to `cmd/main.go`.
- `lib/sort_test.go` benchmarks rebuilt: seeded `rand.NewPCG`, one shared base permutation per size (`benchInput`), fresh clone per iteration, `b.Loop`, `b.ReportAllocs`, pinned `GOMAXPROCS`. Old data race and already-sorted-input measurement bug gone, duplicate/dead stubs gone.
- `TestSelectStepSedgewick` / `TestSelectStepHibbard` exist.

**Outstanding, with the phase that clears each:**

| Problem | Where | Cleared by |
| --- | --- | --- |
| `ShellSort([]int) []int` is `int`-only, non-configurable | `lib/sort.go:12` | Phase 4 |
| `TODO: const for choosing select func` unresolved; gap sequence picked by editing a line | `lib/sort.go:9`, `:15` | Phase 4 |
| `make([]int, 10)` pre-size leaves trailing zeros for short sequences | `lib/sort.go:39`, `:64` | Phase 1 |
| ~~`selectStepHibbard` dead outside its own test~~ | ~~`lib/sort.go:59`~~ | **done in Phase 2** |
| Commented bubble sort + trim lines + duplicate swap | `lib/sort.go:55`, `:76`, `:87-94`, `:103-104` | Phase 4 |
| `d*3 < length` silently drops largest gaps | `lib/sort.go:41`, `:47` | Phase 7 |
| `if step < 5` parallelizes exactly where parallelism is scarcest | `lib/sort.go:19`, `:81` | Phase 7 |
| `TestShellSort` asserts vs hardcoded literals; calls `ShellSort` twice on the same slice (second assert trivially true) | `lib/sort_test.go` | Phase 5 |
| Hardcoded fixtures | `lib/slices.go` | Phase 5 (use), Phase 8 (delete) |
| Package named `lib` in a subdirectory | — | Phase 8 |

Goals: (1) modernize to current Go idioms, (2) make gap sequence a first-class runtime choice with a catalog, (3) build a benchmark study that answers which sequence is best *for this implementation*, (4) restructure per Go library conventions.

Confirmed with user: configurability via **public functional options**; package moves to **repo root**, demo to `examples/`; Hibbard kept as a **selectable alternative**.

---

## Phase 0 — Baseline & prep

Nothing to build; capture the "before" so later phases are attributable.

- [x] `go build ./... && go vet ./...` clean
- [x] `go test -race ./...` green — `ok github.com/OlegElizarov/Shell_Sort_Golang/lib 2.094s`
- [x] `golangci-lint run --timeout=5m` — **baseline is `0 issues.`** (golangci-lint 2.12.2); every later phase must hold it
- [x] `go test -bench=. -benchmem -count=10 ./lib/ > docs/bench-baseline.txt` — pre-change numbers for Phase 6/7 comparison. Recorded 2026-08-03, `GOMAXPROCS=4`, `-count=10`, 123 s:

      `benchstat docs/bench-baseline.txt` — `goos: darwin`, `goarch: amd64`, `cpu: Intel(R) Core(TM) i5-5350U @ 1.80GHz`:

      | Benchmark | sec/op | B/op | allocs/op |
      | --- | --- | --- | --- |
      | `ShellSort/1000-4` | 100.1µ ± 13% | 168.0 ± 0% | 4.000 ± 0% |
      | `ShellSort/10000-4` | 1.360m ± 18% | 168.0 ± 0% | 4.000 ± 0% |
      | `ShellSort/30000-4` | 6.079m ± 24% | 329.0 ± 1% | 5.000 ± 0% |
      | `StdSort/1000-4` | 54.95µ ± 46% | 0 | 0 |
      | `StdSort/10000-4` | 736.4µ ± 7% | 0 | 0 |
      | `StdSort/30000-4` | 2.541m ± 8% | 0.5 ± ? | 0 |
      | geomean | 663.3µ | | |

      Three things this baseline establishes, all of which shape later phases:

      1. **`ShellSort` is ~1.85× slower than `slices.Sort` at `N = 10 000` and ~2.4× at 30 000.** Phase 6's `Sort/std` line has a real gap to close, not a formality.
      2. **Variance is too high for A/B comparison as-is** — ±13–24% on `ShellSort` vs ±7–8% on `StdSort` at the same sizes. The excess is goroutine scheduling noise on a 2-core/4-thread machine, which is exactly what Phase 7 targets. Phase 6/7 need `-count` well above 10, or a quieter machine, before benchstat's p-values mean anything at these spreads.
      3. **`StdSort/1000` at ±46% is a harness artifact, not the sort.** At 55 µs/op the per-iteration `b.StopTimer()`/`b.StartTimer()` toggle in `benchmarkSort` (`lib/sort_test.go:116-118`) is a large fraction of the measurement — the exact defect Phase 6's measurement-hygiene task removes.

      Allocations come from the gap slice plus goroutine machinery; the jump 4→5 at 30 000 is one extra gap.
- [x] `lib.test` (6.9 MB stale binary at repo root) and `coverage.out` — deleted from disk, and covered going forward by `.gitignore` `*.test` / `*.out`

**Gate:** tree builds, tests pass, baseline benchmark file exists.

---

## Phase 1 — Gap catalog, in place

`git commit -m "feat(gaps): add gap sequence catalog"`

Add the catalog to the **existing `lib` package** — no API change yet. Fastest path to a testable artifact, and it flushes out formula bugs while the sources are fresh.

New file `lib/gaps.go`:

- [x] Define the named-value type — a bare func can't be labelled in benchmark output, and several sequences carry provenance worth surfacing:
      ```go
      type GapSequence struct {
          Gaps func(length int) []int // ascending, gaps[0] == 1, all < length
          Name string
      }
      ```
      Field order is `Gaps` before `Name` because govet's `fieldalignment` check (on via `.golangci.yml`) rejects the other order.
      `length` is required, not incidental: Shell, Frank–Lazarus and Gonnet–Baeza-Yates derive gaps *from* `N`; Ciura's fixed tables need `N` to decide where to start extending.
- [x] **Contract every generator obeys** (enforced in Phase 2): ascending, `gaps[0] == 1`, strictly increasing, every element `< length`, never empty (`length < 2` still yields `[]int{1}`).
- [x] Implement the recurrence family: `Tokuda`, `Knuth`, `Hibbard`, `PapernovStasevich`, `Sedgewick86`, `Sedgewick82`
- [x] Implement the closed-form/float family: `Lee`, `SkeanB` (both need `math.Pow` — note that in the doc comments for anyone avoiding float)
- [x] Implement the Pratt family: `Pratt` (2^p·3^q), `Pratt25`, `Pratt34` — generated by merging two geometric progressions and sorting, *not* by a recurrence. Count is `Θ(log²N)`; the only family whose slice is worth pre-sizing.
- [x] Implement the table family: `Ciura`, `Ciura128`, `Ciura1000` — each needs a documented extension rule past its last tabulated term. `⌊2.25h⌋` is Skean et al.'s convention and is **unvalidated**; say so in the doc comment.
- [x] Implement the `N`-derived controls: `IncerpiSedgewick`, `FrankLazarus`, `GonnetBaezaYates`, `Shell`
- [x] All generators `append`-only — no `make([]int, 10)` pre-sizing. Kills the zero-padding footgun at `lib/sort.go:39` and `:64` for good.

Full table with formulas, first terms and roles: [Appendix A](#appendix-a--gap-sequence-catalog). Analysis and citations per sequence: [GAP_SEQUENCES.md §4](GAP_SEQUENCES.md).

**Note on the four dominated sequences.** GAP_SEQUENCES.md §7 says "do not add" Shell, Frank–Lazarus, Papernov–Stasevich, Gonnet–Baeza-Yates. That is advice about *defaults*, not about the catalog: Phase 6's study needs a floor to measure against, and `Shell`/`GonnetBaezaYates` are the only two that produce genuinely bad sequences. Ship them, doc-comment them as controls, and don't recommend them.

**Gate:** `go build ./... && go vet ./...` clean. No behaviour change to `ShellSort` yet.

---

## Phase 2 — Contract + golden-terms tests

`git commit -m "test(gaps): contract and golden-term tests for the catalog"`

New file `lib/gaps_test.go`. This is where the catalog becomes trustworthy — do not carry an untested generator into Phase 3.

- [x] Registry slice `allSequences []GapSequence` so every test loops over the catalog rather than naming entries one at a time
- [x] `TestGapSequenceContract` — for each entry × each `N` in `{0, 1, 2, 16, 128, 1000, 50000}`: ascending, strictly increasing, `gaps[0] == 1`, all `< N` (for `N > 1`), non-empty
- [x] `TestGapSequenceGoldenTerms` — assert the exact first terms from [Appendix A](#appendix-a--gap-sequence-catalog). Transcribed from primary sources; the only defense against a plausible-looking formula that silently generates the wrong sequence, which is the failure mode that cost the most time while writing GAP_SEQUENCES.md
- [x] Delete `TestSelectStepSedgewick` / `TestSelectStepHibbard` — superseded, and their `rand.IntN` lengths are non-reproducible
- [x] **Pulled forward from Phase 4:** deleting those tests left `selectStepHibbard` with no callers, and `unused` flagged it — holding the zero-lint baseline required removing the function and its commented call site (`lib/sort.go:16`) now rather than in Phase 4. Its replacement `Hibbard` was already in the catalog, so nothing was lost
- [x] `TestCiuraExtension` (**added, not in the original plan**) — the `⌊2.25h⌋` continuation past each Ciura table is the one part of the catalog with no source to check against, so it gets its own test: the last tabulated term is present, extension actually happens, and every extended term is `⌊2.25·previous⌋`
- [x] `t.Parallel()` at top level and in every subtest (matches existing house style; `paralleltest` linter enforces it)
- [x] Update this file's progress table

**Gate:** all 18 sequences pass contract + golden terms under `go test -race ./...`.

---

## Phase 3 — Counting harness + literature validation

`git commit -m "test(bench): instrumented counting harness validated against published counts"`

**The gate on the whole catalog.** Cheapest correctness check available for Phase 1's 18 generators. Do not proceed past a mismatch.

- [x] Add unexported `sortInstrumented` in `lib/gaps_bench_test.go` mirroring `subSort` with two counters — test files are excluded from the shipped package, so instrumentation never touches the production path
- [x] Match Skean et al.'s definitions exactly or the numbers aren't comparable: a **comparison** is one evaluation of `A(i) > A(i+k)`; an **exchange** is one swap performed to fix an inversion
- [x] The current `subSort` uses a *shifting* insertion sort — decide explicitly whether a shift counts as an exchange, and record the choice in a comment next to the counter
- [x] Run Tokuda, Ciura-Large, Pratt-23, Pratt-25 at `N = 10 000`, mean over ~~1000~~ **200** random permutations (seeded `rand.Perm`). 200 was chosen over the published 1000 because the per-permutation spread is small enough that 200 pins the mean far tighter than the 2% tolerance, and it keeps the test runnable under `-race` in CI — the full suite runs in ~8 s. The test skips under `-short`
- [x] Compare against the published values:

      | Sequence | Expected μCO | Expected μEX |
      | --- | --- | --- |
      | Tokuda | 192 574 | 98 071 |
      | Ciura (Large) | 191 435 | 101 680 |
      | Pratt-23 | 604 502 | 66 923 |
      | Pratt-25 | 450 131 | 62 191 |

      Counts are implementation-independent given matching definitions, so agreement within a percent or two validates both the generators and the counter. **Disagreement means a bug in the catalog, not a discovery.**

      **Result: 7 of 8 reproduced within 0.27%.** Measured (200 trials, `N = 10 000`):

      | Sequence | μCO measured | Δ | μEX measured | Δ |
      | --- | --- | --- | --- | --- |
      | Tokuda | 192 482 | −0.048% | 98 176 | +0.107% |
      | Ciura-Large | 191 382 | −0.028% | 101 840 | +0.158% |
      | Pratt-23 | 604 438 | −0.011% | **63 705** | **−4.81%** |
      | Pratt-25 | 449 982 | −0.033% | 62 357 | +0.267% |

      The one miss is Pratt-23's exchange count, and the rule above was applied rather than waived — the deviation was investigated, and the evidence exonerates the catalog. Pratt-23's *comparisons* reproduce to 0.011%, and comparisons depend on every one of the 67 gaps generated at this size, so a wrong gap set cannot match them to four decimals while missing only exchanges. Capping the generator at `N/2`, `N/3` or `N/4` moves exchanges toward 66 923 but destroys the comparison agreement (580 166 / 553 377 / 534 384 vs 604 502), so no single gap set produces both published figures. Pratt-25 runs the identical code path with a different base and reproduces both. The test asserts the measured 63 705 at the same tolerance, so it still guards regressions.
- [x] Record the reproduced numbers next to the published ones in GAP_SEQUENCES.md §5.1
- [x] `TestCountingSortSorts` (**added, not in the original plan**) — the counts only mean something if the mirrored algorithm sorts, so the harness is checked against a `slices.Sort` oracle across all 18 sequences

**Gate:** ✅ passed. Seven of eight reproduce within 0.27%; the eighth is a published-table discrepancy, argued above and documented at `lib/gaps_bench_test.go`'s `prattMeasuredEX`, not a catalog bug.

> Phases 1–3 are worth doing even if the rest is deferred: they turn the literature survey into executable, verified code.

---

## Phase 3.1 — Parked nits and small fixes

**Not a gate, and not sequenced here.** This is the backlog for small improvements raised mid-flight. Each item names the phase that should absorb it, so nothing gets done twice or done too early — landing them where the surrounding code is already being edited keeps the diffs attributable.

### 3.1.1 — Numeric sequence identity instead of a bare `Name` string

**Absorb into:** Phase 4 (it changes the public type) · **Status:** ⬜ open

Give each catalog entry a generated enum identity rather than carrying a `string` as its only handle:

```go
//go:generate stringer -type=SequenceID
type SequenceID uint8

const (
    TokudaID SequenceID = iota
    CiuraID
    // …
)
```

The honest case for it, strongest first:

1. **`GapSequence` is not comparable today.** It has a `func` field, so `==`, map keys and `slices.Contains` are all unavailable — the reason `TestGapSequenceGoldenTerms` matches cases to the catalog by position and a length assertion rather than by identity. A `SequenceID` is comparable, map-keyable and sortable, which is what a registry, a CLI flag, or a `BENCHMARKS.md` result table each want.
2. **Exhaustiveness.** With an enum, `go vet`/`exhaustive` can flag a catalog entry that was added but never registered. Right now that is caught only by the hand-written `require.Len(cases, len(allSequences))` guard.
3. **`stringer` gives the label back for free**, so the struct can drop `Name` and shrink — which also removes the `fieldalignment` workaround noted in Phase 1.

**On the performance argument specifically:** numeric comparison does beat string comparison, but it is worth being precise about where that would pay. There is no string comparison on the sorting path today — `WithGapSequence` will take a `GapSequence` value, and `Name` is only read when a test or benchmark formats a label. So the win is *structural* (comparability, exhaustiveness, registry lookups) rather than measurable in `ns/op`. If a future API does dispatch by name — a `--gap-sequence=tokuda` flag, a `map[string]GapSequence` lookup per call — then the numeric form becomes a real cost argument too, and an `O(1)` array index by ID beats a map hash either way.

**Cost:** a generated `sequenceid_string.go` plus a `stringer` dependency (`golang.org/x/tools/cmd/stringer`) in the toolchain, and CI needs `go generate` output checked in or verified.

### 3.1.2 — Retrospective: measured improvement per phase

**Absorb into:** Phase 6 (build it) and Phase 7 (first real entry) · **Status:** ⬜ open

Track what each phase actually bought, rather than only reporting the final number. `docs/bench-baseline.txt` from Phase 0 is the first data point; the intent is a performance changelog that answers "which change earned the speedup".

Shape:

- One recorded artifact per phase that can move performance — `docs/bench-phase{N}.txt` — produced the same way every time (same `-count`, same machine, same `GOMAXPROCS`), so `benchstat docs/bench-baseline.txt docs/bench-phase7.txt` is always valid.
- A `make bench-record PHASE=n` target so the recording procedure is not retyped and cannot drift.
- A cumulative table at the top of `docs/BENCHMARKS.md`: phase, commit, what changed, sec/op at each size, delta vs the previous phase, delta vs baseline, and delta vs `slices.Sort` — the last column being the one that says whether the library is worth using at all.

**The measurement caveat this has to respect:** Phase 0 recorded ±13–24% variance on `ShellSort` on this machine, so a phase-to-phase delta smaller than roughly 25% is not distinguishable from noise at `-count=10`. Either raise `-count` substantially for recorded runs or report benchstat's p-value alongside each delta and treat anything above 0.05 as "no measured change". A retrospective table that presents noise as progress is worse than no table.

### 3.1.3 — *(awaiting the third item)*

---

## Phase 4 — Options API + generics

`git commit -m "feat(sort): generic ShellSort with functional options"`

Now the public API changes. `lib/sort.go` stays in place — the move is Phase 8, so no behavioural change hides inside a rename.

- [ ] Generic signature matching stdlib's shape; type inference keeps existing int-slice call sites compiling unchanged:
      ```go
      func ShellSort[S ~[]E, E cmp.Ordered](x S, opts ...Option) S
      func subSort[T cmp.Ordered](x []T, gap, start int) // no pointer, no *sync.WaitGroup param
      ```
- [ ] Functional options, resolving the TODO at `lib/sort.go:9`:
      ```go
      type Option func(*config)
      func WithGapSequence(g GapSequence) Option
      func WithParallelThreshold(minSubarrayLen int) Option // wired up in Phase 7
      ```
- [ ] Default is `Tokuda` — closed form, unbounded, ~17 passes at `10^6`, ties every Ciura variant on comparisons at `N = 10^4` while beating all of them on exchanges, and its `N`-scaling first gap matters most for the parallel structure (GAP_SEQUENCES.md §6.3, §7.1)
- [ ] Drop `selectStepSedgewick` — `selectStepHibbard` already went in Phase 2 when `unused` flagged it. Both now live in the catalog as `Sedgewick86` / `Hibbard`, per the user's decision to keep Hibbard selectable rather than delete it
- [ ] Remove dead code: commented bubble sort (`lib/sort.go:87-94`), commented trim lines (`:55`, `:76`), commented duplicate swap (`:103-104`)
- [ ] `subSort` takes no `*sync.WaitGroup` — the `if step < 5 { defer wg.Done() }` coupling at `lib/sort.go:81` is what makes the current signature load-bearing; caller owns the barrier
- [ ] Doc comment on every exported identifier — preserves the zero-lint baseline
- [ ] Update `cmd/main.go` if the call site needs it (it shouldn't — inference covers it)

**Gate:** `go build ./... && go vet ./...` clean; existing tests still pass unmodified; `golangci-lint` still zero.

---

## Phase 5 — Test rebuild

`git commit -m "test(sort): oracle-based table tests, all-sequences loop, fuzz target"`

- [ ] Shared seeded generator replacing the fixtures:
      ```go
      func randomSlice(tb testing.TB, n int, seed1, seed2 uint64) []int {
          tb.Helper()
          return rand.New(rand.NewPCG(seed1, seed2)).Perm(n)
      }
      ```
- [ ] Rewrite `TestShellSort` — table of `{Name, Size}` (empty, single, 16, 100, 500, plus explicit already-sorted and reverse-sorted)
- [ ] Oracle is `slices.Sort` on a clone, not a literal: `want := slices.Clone(in); slices.Sort(want)` vs `got := ShellSort(slices.Clone(in))`. Explicit cloning fixes the in-place-mutation bug where the current test's second assertion is trivially true
- [ ] `TestShellSortAllSequences` — run the correctness table against **every** catalog entry. Payoff of Phase 1's contract: one loop covers 18 implementations
- [ ] `FuzzShellSort` (stdlib `testing.F`) — fuzz `(seed int64, n uint8)`, bounded value range so duplicates and ties appear (which `Perm` structurally cannot produce), assert against the `slices.Sort` oracle. Seed corpus includes empty and single-element. Runs under plain `go test ./...`
- [ ] Keep `lib/slices.go` for now — `cmd/main.go` still imports it; deletion is Phase 8

**Gate:** `go test -v -race ./...` green, including the fuzz seed corpus.

---

## Phase 6 — Benchmark matrix

`git commit -m "test(bench): full sequence × size × shape matrix"` + `docs: add measured benchmark results`

The central question — which sequence is fastest *here* — is not answerable from the literature. Published rankings are dominated by comparison counts measured in Python, and the top group is separated by under 2% (GAP_SEQUENCES.md §5.1), well inside what goroutine overhead and cache behaviour will swamp in this implementation.

**Three metrics, measured separately. They disagree, and that disagreement is the finding:**

1. **Comparisons** — literature-comparable, implementation-independent (harness from Phase 3)
2. **Exchanges** — Pratt beats Tokuda by ~30% here while losing 3.1× on comparisons
3. **Wall-clock + allocs/op** — the only metric that decides the default

- [ ] `Sort/{sequence}/n={size}` across the full catalog, `benchSizes = {1_000, 10_000, 100_000}`
- [ ] `Sort/std/n={size}` — `slices.Sort` baseline on the identical base permutation
- [ ] `Sort/{parallel,sequential}/{best sequence}/n={size}` — quantifies the goroutine machinery against `WithParallelThreshold` disabled. GAP_SEQUENCES.md §6 predicts a modest ceiling; this measures it
- [ ] Input shapes beyond random permutations: already-sorted, reverse-sorted, few-unique. Weiss's `O(N log N)` result for reverse-ordered input means the adversarial case for plain insertion sort is *not* adversarial for Shellsort — the suite should show that
- [ ] **Measurement hygiene:** pre-clone all iterations' inputs before `b.ResetTimer()` rather than calling `b.StopTimer()`/`b.StartTimer()` inside the loop — the current helper pays timer-toggle overhead every iteration, significant at these sizes
- [ ] Run `make bench-record && make bench-stat` (defaults to `-count=20` against `docs/bench-baseline.txt`). **`-count=10` is not enough here** — Phase 0 measured ±13–24% spread on `ShellSort`, and a 3-count control run against that baseline with *zero code changed* reported `ShellSort/30000-4  -27.05% (p=0.028)`. Check the reported `±` before trusting any delta; a p-value under a wide spread is not a result
- [ ] Write `docs/BENCHMARKS.md` — the empirical answer belongs next to the literature survey
- [ ] Flip the default in Phase 4's config if the data disagrees with Tokuda, and say why in `BENCHMARKS.md`

**Gate:** `docs/BENCHMARKS.md` exists with benchstat output for the full matrix.

---

## Phase 7 — Parallel redesign

`git commit -m "perf(sort): chunk goroutines and thread parallelism by work per subarray"`

Separate commit from Phase 6 so the effect is attributable. Two constructs in the current code are unjustified, and GAP_SEQUENCES.md §6 explains why.

- [ ] Replace `if d[...] < 5` (`lib/sort.go:19`) — it parallelizes only when there are *fewer* than 5 subarrays, i.e. exactly when there is least parallelism available and each subarray is longest, so the wide early passes run sequentially. Threshold on **work per subarray** instead (`len(x)/gap >= minSubarrayLen`), so parallelism follows the work rather than the subarray count. Wire it to `WithParallelThreshold`
- [ ] Remove `d*3 < float64(length)` (`lib/sort.go:41`, `:47`) from the shared path — an undocumented cap that silently drops the largest gaps. Knuth's rule is to stop at `⌈N/3⌉`, which is a *stopping condition for the generator*, not a filter applied to every element. Fold it into the `Knuth` generator where it belongs
- [ ] **Decouple goroutine count from gap count.** One goroutine per subarray means the first pass of a closed-form sequence spawns thousands of goroutines each sorting a handful of elements. Chunk instead: partition the `gap` subarrays across `min(gap, GOMAXPROCS)` workers
- [ ] Re-run the Phase 6 matrix, before/after, as its own line in the study — this is likely the single largest available win and is independent of sequence choice
- [ ] `go test -race ./...` — the chunking rewrite is the highest-risk change in the plan for data races
- [ ] Record results in `docs/BENCHMARKS.md`

**What will *not* improve, and should be stated in the doc rather than chased:** the final `h = 1` pass is a single subarray of `N` elements and is unparallelizable, so span is dominated by the low-width tail passes regardless of sequence or chunking. Amdahl caps the whole approach. Sequence choice changes barrier count (~17 for Tokuda vs ~125 for Pratt at `N = 10^6`), not the floor.

**Gate:** benchstat before/after in `BENCHMARKS.md`; race detector clean.

---

## Phase 8 — Restructure to root package

`git commit -m "refactor: move package to repo root as shellsort"`

Pure moves, last, so no behavioural change hides inside a rename. Ideally a `git mv`-only diff plus import-path edits.

- [ ] `lib/sort.go` → `sort.go`, `package lib` → `package shellsort`; `lib/gaps.go` → `gaps.go`
- [ ] `lib/sort_test.go` → `sort_test.go` + `sort_bench_test.go` + `fuzz_test.go`; `lib/gaps_test.go` → `gaps_test.go`; `lib/gaps_bench_test.go` → `gaps_bench_test.go`
- [ ] Delete `lib/slices.go` and the now-empty `lib/`
- [ ] `cmd/main.go` → `examples/shellsort/main.go`; import path becomes `github.com/OlegElizarov/Shell_Sort_Golang` aliased `shellsort` (last path segment won't lexically match the package name — normal in Go). Demo builds a seeded slice inline instead of importing the removed fixtures
- [ ] Update `Makefile`: `BENCH_PKG ?= ./lib/...` → `.`, and `test-short` / `demo` paths. **This is the one thing the restructure breaks that the compiler won't catch**
- [ ] No changes to `.github/workflows/ci.yml` or `go.mod` — CI globs `./...`, the benchmark smoke step goes through `make bench-smoke` (package path lives in the Makefile), module path unchanged. Verify rather than assume
- [ ] `go build ./... && go test -race ./... && go run ./examples/shellsort`

**Gate:** tests pass with no edits to test *logic* — only package clause and imports.

---

## Phase 9 — Docs

`git commit -m "docs: update README and CLAUDE.md for the new layout"`

- [ ] `README.md` — import path, `go run ./examples/shellsort`, sequence catalog, link to `BENCHMARKS.md`
- [ ] `CLAUDE.md` — paths, package name, resolved TODO, new commands. Note it currently claims `main.go` is at the repo root, which is **already stale**
- [ ] GAP_SEQUENCES.md §5.1 — measured results alongside the published ones
- [ ] Mark this plan complete; move it to `docs/archive/` or leave with all boxes ticked

**Gate:** `gopls check ./...` — zero diagnostics, confirming doc-comment coverage on the ~20 new exported identifiers.

---

## Verification (full run, after Phase 9)

- [ ] `go build ./...`, `go vet ./...` — clean across the new layout
- [ ] `go test -v -race ./...` — table tests, all-sequences loop, contract + golden terms, fuzz seed corpus, no benchmark data race
- [ ] `go test -bench=. -benchmem ./...` — full matrix across sizes, sequences, input shapes, parallel/sequential
- [ ] Literature reproduction (Phase 3) still passes
- [ ] `golangci-lint run --timeout=5m` — zero new diagnostics vs the Phase 0 baseline
- [ ] `gopls check ./...` — zero diagnostics
- [ ] `go run ./examples/shellsort` — demo still prints sorted output

---

## Appendix A — Gap sequence catalog

Implemented in Phase 1. Each is a handful of lines; the point of having them all is that Phase 6's study is only meaningful across the full field, including the known-bad ones as controls. Analysis and citations: [GAP_SEQUENCES.md](GAP_SEQUENCES.md).

Two rows below were corrected during Phase 1: `Pratt25`'s terms had listed 15 (= 3·5, not `2^p·5^q`) and `Pratt34`'s had listed 24 (= 2³·3, not `3^p·4^q`). Both are now generated from the definitions and verified against the sources.

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
| `Pratt25` | `2^p·5^q` | 1, 2, 4, 5, 8, 10, 16, 20 | best exchanges measured |
| `Pratt34` | `3^p·4^q` | 1, 3, 4, 9, 12, 16, 27 | Pratt family, fewer comparisons |
| `Sedgewick86` | piecewise even/odd `k` | 1, 5, 19, 41, 109 | current default, keep for continuity |
| `Sedgewick82` | `4^k + 3·2^{k−1} + 1` | 1, 8, 23, 77, 281 | ratio-4 datapoint |
| `Hibbard` | `2^k − 1` | 1, 3, 7, 15, 31 | keep per user decision |
| `PapernovStasevich` | `2^k + 1`, prefixed 1 | 1, 3, 5, 9, 17, 33 | ratio-2 control |
| `IncerpiSedgewick` | constructed, coprime triangle | 1, 3, 7, 21, 48, 112 | best asymptotic exponent |
| `FrankLazarus` | `2⌊N/2^{k+1}⌋ + 1` | …, 3, 1 (`N`-derived) | `N`-dependent control |
| `GonnetBaezaYates` | `max(⌊(5h−1)/11⌋, 1)` from `h_0 = N` | 1, 3, 8, 19, 42, 93, 206, 454 (`N`=1000) | ratio-2.2, `Θ(N²)` for some `N` |
| `Shell` | `⌊N/2^k⌋` | N/2, N/4, …, 1 | negative control (`Θ(N²)`) |

## Appendix B — Critical files

| Path | Fate | Phase |
| --- | --- | --- |
| `lib/sort.go` | → `sort.go`, generic + options | 4, 7, 8 |
| `lib/gaps.go` | new, then → `gaps.go` | 1, 8 |
| `lib/gaps_test.go` | new, then → `gaps_test.go` | 2, 8 |
| `lib/gaps_bench_test.go` | new (counting harness) | 3 |
| `lib/sort_test.go` | → `sort_test.go` + `sort_bench_test.go` + `fuzz_test.go` | 5, 6, 8 |
| `lib/slices.go` | deleted | 8 |
| `cmd/main.go` | → `examples/shellsort/main.go` | 8 |
| `README.md`, `CLAUDE.md` | updated | 9 |
| `docs/GAP_SEQUENCES.md` | measured results added | 3, 9 |
| `docs/BENCHMARKS.md` | new | 6, 7 |
| `docs/bench-baseline.txt` | new | 0 |
