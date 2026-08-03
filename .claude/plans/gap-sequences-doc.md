# Plan: gap sequence analysis doc

Status: executed 2026-07-30. Deliverable is `docs/GAP_SEQUENCES.md`.

## Context

`lib/sort.go:15` hardcodes `selectStepSedgewick`, with `selectStepHibbard` commented out one line below — the only two gap sequences the repo knows about, and `README.md` documents them as the whole menu. Before `docs/MODERNIZATION_PLAN.md` §A makes the gap function pluggable via `WithGapSequence`, there was no written basis for deciding *which* sequences are worth adding: no formulae, no complexity bounds, no comparison against this repo's one-goroutine-per-subarray design.

Scope decided with user:
- **Literature analysis only** — no benchmarks run, no code changed.
- Full survey (~12 sequences) **plus** per-sequence fit with the repo's parallel structure.
- Parallel analysis assumes **enough CPUs for every subarray**. The `< 5` goroutine threshold (`lib/sort.go:19`) and the `d*3 < length` guard (`lib/sort.go:41`) are explicitly **out of scope** — possible flaws in the current implementation, not constraints to design around.

## Document structure (`docs/GAP_SEQUENCES.md`)

1. **Notation** — `h_1 = 1 < ... < h_t` applied decreasing; pass with gap `h` = `h` subarrays of length ~`N/h`; `h`-sortedness preservation; Frobenius argument; `t(N) ≈ log_r N`.
2. **Metrics scored** — formula kind, ratio `r`, pass count, worst case, lower bound, average case (flagged measured vs proven), adaptivity, locality, gap-table cost, parallel width/span/barriers. Noted as universal: unstable, in-place.
3. **Theory anchors** — Pratt `Θ(N log²N)`; Plaxton–Poonen–Suel `Ω(N(log N/log log N)²)`; Jiang–Li–Vitányi `Ω(pN^{1+1/p})`; Knuth two-gap `2N²/h + √(πN³h)`; Janson–Knuth `O(N^{23/15})`; Weiss reverse-order `O(N log N)`; coprimality principle; Zang 2026 `Ω(N^{1.26})` for Tokuda and all rational-geometric families.
4. **The sequences** — Shell, Frank–Lazarus, Papernov–Stasevich, Pratt, Knuth, Incerpi–Sedgewick, Sedgewick 1982, Gonnet–Baeza-Yates, Tokuda, Ciura, Lee 2021, Skean et al. 2023, plus a ratio-family subsection (`r` = 2, 2.2–2.3, 3, 4, φ). Hibbard + Sedgewick 1986 as baselines only.
5. **Summary table** — formula, first terms, `r`, `t(10^6)`, worst case, average case, notes.
6. **Repo parallel fit** — three findings: (a) `h = 1` pass is an `Ω(N)` serial floor and span is dominated by the low-width tail passes, so Amdahl caps speedup for every sequence; (b) pass count == barrier count, so ratio ~2.25 wins twice and Pratt's ~125 barriers make its best-in-class bound useless here; (c) Ciura's fixed 701 ceiling caps parallel width and loads the first pass — argues for Tokuda over Ciura more strongly than the sequential literature does.
7. **Recommendation** — Tokuda default, Ciura opt-in (small `N`, extension flagged unproven), Knuth fallback, Lee experimental, Pratt docs-only, reject Shell/Frank–Lazarus/Papernov–Stasevich/Gonnet–Baeza-Yates.
8. **References** — primary sources for every sequence and bound.

## Accuracy rule applied

Per `feedback_no_hallucinate_versions`, every constant and bound was verified against a source before writing, not recalled. Corrections this caught:

- "Gonnet & **Baase**" → **Gonnet & Baeza-Yates**; formula is top-down `max(⌊(5h−1)/11⌋, 1)` from `h_0 = N`, and its worst case is `Θ(N²)` for certain `N` — disqualifying, which the first draft had missed.
- Tokuda's closed form → `⌈h'_k⌉` with `h'_k = 2.25h'_{k-1} + 1`, `h'_1 = 1` (several floating-point variants circulate).
- Lee 2021 → `⌈(γ^k − 1)/(γ − 1)⌉`, γ = 2.243609061420001, terms `1, 4, 9, 20, 45, 102, 230, 516`.
- Worst-case lower bound attribution → Plaxton–Poonen–Suel, not "Poonen".
- Added Zang 2026 (arXiv:2607.08997), which caps the entire practical ratio family at `Ω(N^{1.26})` — flagged as unrefereed preprint.

## Files

- `docs/GAP_SEQUENCES.md` — new, the deliverable.
- `README.md` — link added under the existing `## Gap Sequences` section.
- No changes to `lib/`, `cmd/`, `CLAUDE.md`, or CI.

## Verification

- `go build ./...` + `go test ./...` clean and unchanged — confirms doc-only change.
- Formulae cross-checked against primary sources (Wikipedia's gap table used only as a cross-check index, cited as such).
- Pass counts are `log_r 10^6`, Pratt `(log₂N)(log₃N)/2`.
