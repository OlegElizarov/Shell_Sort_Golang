# Shellsort Gap Sequences: Analysis Beyond Sedgewick and Hibbard

Literature survey. No benchmarks were run for this document — every number here is either a proven bound or a published measurement, attributed as such.

`lib/sort.go` currently ships two gap sequences: `selectStepSedgewick` (active) and `selectStepHibbard` (commented out). This document surveys everything else, so that making the gap function pluggable (see `docs/MODERNIZATION_PLAN.md` §A, `WithGapSequence`) becomes a decision with evidence behind it. Sedgewick and Hibbard appear only as baselines in the comparison table.

## 1. Notation

A gap sequence is `h_1 = 1 < h_2 < ... < h_t`, applied in **decreasing** order. A pass with gap `h`:

- splits the array into `h` interleaved subarrays, subarray `i` being indices `i, i+h, i+2h, ...`;
- each subarray has length ~`N/h`;
- insertion-sorts each subarray, leaving the array **`h`-sorted**.

Correctness requires only `h_1 = 1` (the final pass is a plain insertion sort). Everything else is performance.

Two properties do the theoretical work:

- **`h`-sortedness is preserved** by later passes with smaller gaps. Work is never undone.
- Passes **inherit** work from earlier passes, but only when the gaps are arithmetically compatible. This is the whole ballgame, and it is worked out in §1.1.

`t(N)` = number of passes = number of gaps below `N`. For a sequence with asymptotic ratio `r = h_{k+1}/h_k`, `t(N) ≈ log_r N`.

### 1.1 Why gap arithmetic decides everything

The single mechanism behind every good and every bad sequence in this document.

**Setup.** Say the array is already `k`-sorted and `l`-sorted — two earlier passes, with gaps `k` and `l`, have finished. Now a pass with gap `h` begins. How much work is left for it?

**The chain argument.** Suppose `h` can be written as `h = a·k + b·l` with `a, b` non-negative integers — `a` hops of size `k` plus `b` hops of size `l` land you exactly `h` positions along. Because the array is `k`-sorted, every `k`-hop goes non-decreasing; same for every `l`-hop. Chain them:

```
A[i] ≤ A[i+k] ≤ A[i+2k] ≤ ... ≤ A[i+ak] ≤ A[i+ak+l] ≤ ... ≤ A[i+ak+bl] = A[i+h]
```

So `A[i] ≤ A[i+h]` **already holds before the `h`-pass starts**. Every distance representable as `ak + bl` is pre-sorted for free.

**Which distances are representable?** This is the coin problem: given coins of denominations `k` and `l`, which totals can you make? If `k` and `l` are coprime, everything above `kl − k − l` (the **Frobenius number**) is reachable, and only finitely many values below it are not. Concrete: `k = 2`, `l = 3` gives `kl − k − l = 1`, so *every* distance except 1 is representable.

**Why that yields `O(N)` per pass.** Only the finitely many non-representable distances can still be out of order. Each element in the `h`-pass therefore has a bounded window of possible offenders, independent of `N` — so its insertion moves it a constant number of positions, and `N` elements cost `O(N)`. The pass is linear no matter how large the array.

**Pratt, concretely.** Gaps are the 3-smooth numbers `2^p·3^q`. When the `h`-pass runs, `2h` and `3h` are also 3-smooth and larger, so both already ran. Take `k = 2h`, `l = 3h`:

```
a·2h + b·3h = h·(2a + 3b),  and 2a + 3b covers 2, 3, 4, 5, 6, ...
```

Every multiple of `h` is covered except `1·h` itself. Each element can be out of place by at most one position, so each pass costs `O(N)` — and `Θ(log²N)` passes give the `Θ(N log²N)` bound. Pratt's sequence is not fast by accident; it is constructed so that every gap is spanned by two gaps that preceded it.

**Why shared factors are fatal.** If `k` and `l` share a common factor `d`, then `a·k + b·l` is always a multiple of `d`. Distances outside that lattice are never reachable — the Frobenius number is infinite, and no amount of earlier work transfers.

Shell's original `..., 8, 4, 2, 1` is the worst case of this. When the 2-pass runs, the completed passes were 8 and 4, and `a·4 + b·8` produces only multiples of 4. Distances 2, 6, 10, … inherit **nothing**. Every pass starts from scratch, no pass is cheap, and the final `h = 1` pass — a full insertion sort on data that earlier passes barely improved — does essentially all the work. Hence `Θ(N²)`, no better than skipping Shellsort entirely.

Compare Hibbard's `2^k − 1`: consecutive gaps like 7 and 15 are coprime, the lattice fills in, and the bound drops to `Θ(N^{3/2})`. The sequences differ only in arithmetic, not in structure.

**The design rule this implies**, and the reason for the `+1` and `−1` correction terms scattered through Hibbard, Papernov–Stasevich, Knuth and Sedgewick: consecutive gaps must be coprime, or nearly so. A bare `⌈r^k⌉` with integer `r` reintroduces Shell's failure mode exactly.

### Symbols used in the measurement tables

Every symbol that appears in a results table in this document or in `docs/BENCHMARKS.md`, spelled out once:

| Symbol | Reads as | What it actually counts |
| --- | --- | --- |
| `N` | input size | How many elements are being sorted. Results are always quoted at a stated `N`, because every sequence's ranking depends on it. |
| **CO** | comparisons | One evaluation of "is `A(i)` greater than `A(i+h)`?" — one question the algorithm asks about two elements. Includes the question that comes back "no" and ends the inner loop. |
| **EX** | exchanges | One swap that fixes one inversion, i.e. one pair of elements put back in the right order relative to each other. |
| **μ** | mean (the Greek letter mu) | An average over many runs. `μCO` is therefore "average number of comparisons", and `μEX` is "average number of exchanges" — averaged over some number of randomly shuffled inputs, because a single input's count depends on how unlucky that particular shuffle was. |
| **trials** | — | How many random shuffles were averaged to get a `μ`. More trials, steadier average. |
| **Δ** | delta, i.e. difference | How far a measured number sits from the one it is being checked against, in percent. `−0.048%` means the measured value is 0.048% *below* the published one. |

Why these two metrics and not just "speed": comparisons and exchanges are **implementation-independent**. Two correct Shellsorts using the same gap sequence perform the same number of comparisons on the same input, whether written in Python or Go, on any machine, at any clock speed. That makes them the only numbers that can be checked against a published paper. Wall-clock time is *not* comparable that way — it belongs in `docs/BENCHMARKS.md`, measured here, on this machine.

The two also disagree about which sequence is best, which is a finding rather than a nuisance: a sequence can ask fewer questions but do more moving, or the reverse. §5.1 shows exactly that.

## 2. Metrics scored

Each sequence below is judged on the same axes:

| Metric | Why it matters |
| --- | --- |
| **Formula** | Closed form (computable for any `N`) vs recurrence vs fixed empirical table (needs an extension rule). |
| **Ratio `r`** | Sets pass count and how fast subarrays get long. Empirically `r ≈ 2.2–2.3` wins. |
| **Pass count `t(N)`** | Each pass is a full sweep over `N` elements — and in this repo, one `sync.WaitGroup` barrier. |
| **Worst case** | Proven upper bound on comparisons. |
| **Lower bound** | What is proven impossible for that sequence — separates "unanalyzed" from "known bad". |
| **Average case** | Almost always empirical. Flagged explicitly when it is a measurement, not a theorem. |
| **Adaptivity** | Behaviour on sorted / reverse-sorted / nearly-sorted input. |
| **Locality** | Large gaps stride past cache lines: one miss per access. Small gaps are sequential. Fraction of work spent in small-gap passes is a cache-cost proxy. |
| **Gap-table cost** | Memory + arithmetic to produce the gaps. Relevant to `make([]int, 10)` pre-sizing at `lib/sort.go:39`. |
| **Parallel width / span / barriers** | See §5. |

Not applicable to any sequence: **stability** (Shellsort is unstable for every gap sequence, because long-range swaps reorder equal keys) and **extra memory** (all variants are in-place, `O(t)` for the gap slice).

## 3. Theory anchors

Bounds that constrain every possible sequence:

- **Pratt 1971** — `Θ(N log²N)` worst case, the best proven for any gap sequence. No sequence is known to beat it asymptotically in the worst case.
- **Plaxton, Poonen & Suel** — worst-case lower bound `Ω(N (log N / log log N)²)` for *any* gap sequence, later refined toward `Ω(N (log N)² / log log N)` under conditions. Pratt is therefore within a `(log log N)²`-ish factor of optimal, and no gap sequence will ever make Shellsort `O(N log N)` in the worst case.
- **Jiang, Li & Vitányi 2000** — average-case lower bound `Ω(p N^{1+1/p})` for a `p`-pass Shellsort, `p ≤ log₂N`. Fewer passes are provably costly: 2 passes cannot beat `Ω(N^{3/2})`, 3 cannot beat `Ω(N^{4/3})`. This is the formal reason a good sequence uses `Θ(log N)` passes.
- **Knuth, two-gap average** — for gaps `(h, 1)` the average running time is `2N²/h + √(πN³h)`, minimized around `h ≈ N^{1/3}` giving `Θ(N^{5/3})`. The cleanest closed-form illustration of the pass-count/pass-cost tradeoff.
- **Janson & Knuth, three-pass** — `O(N^{23/15})` average with optimized gaps.
- **Weiss** — Shellsort runs in `O(N log N)` on reverse-ordered input, i.e. the classic adversarial case for insertion sort is *not* adversarial here.
- **Coprimality principle** (derived in full in §1.1) — gaps sharing a factor `f` inherit no work from each other, because `a·k + b·l` cannot leave the lattice of multiples of `f`. Pure powers of two are the canonical failure. Every good sequence keeps consecutive gaps coprime or nearly so.
- **Zang 2026** (preprint, arXiv:2607.08997, unrefereed — treat as provisional) — first individually proven nontrivial lower bounds for empirically-derived sequences: `Ω(N^{1.26})` worst case for **Tokuda's** sequence, and the same for *any* strictly decreasing sequence approximating a rational geometric progression. Consequence worth internalizing: the whole ratio-`r` family, which contains every practical sequence in §4, is capped at `Ω(N^{1.26})` worst case. Beating that requires Pratt-style structure, not a better constant.

## 4. The sequences

### Shell 1959 — `⌊N/2^k⌋`

Gaps `⌊N/2⌋, ⌊N/4⌋, ..., 1`. Worst case **`Θ(N²)`** — no better than insertion sort. Cause is the coprimality failure above: every gap divides the previous one. `t(N) = log₂N`, ratio 2. Of historical interest only; useful in the doc as the negative control that shows ratio alone doesn't determine quality.

### Frank & Lazarus 1960 — `2⌊N/2^{k+1}⌋ + 1`

Gaps `2⌊N/4⌋+1, 2⌊N/8⌋+1, ..., 3, 1`. Same halving as Shell but forced odd, which breaks the shared factor. Worst case drops to **`Θ(N^{3/2})`** — the single cheapest fix in the entire history of the algorithm, and the clearest demonstration that coprimality, not growth rate, is what matters. `N`-dependent (gaps derived from `N`, not generated bottom-up).

### Papernov & Stasevich 1965 — `2^k + 1`, prefixed with 1

Gaps `1, 3, 5, 9, 17, 33, 65, ...`. Worst case **`Θ(N^{3/2})`**, matching Hibbard with a slightly better constant in practice. Ratio 2, so `t(N) ≈ log₂N` passes — roughly 20 for `N = 10^6`, about 20% more passes than a ratio-2.25 sequence for no gain. Trivially cheap to compute; superseded by Knuth.

### Pratt 1971 — all 3-smooth numbers `2^p · 3^q < N`

Gaps `1, 2, 3, 4, 6, 8, 9, 12, 16, 18, 24, 27, ...`. Worst case **`Θ(N log²N)`** — the best proven bound of any sequence, because after `2h`- and `3h`-sorting, the `h`-pass costs `O(N)` (Frobenius argument with coins 2 and 3).

The catch is the pass count: `t(N) ≈ (log₂N)(log₃N)/2`, i.e. **`Θ(log²N)`**, about 125 passes at `N = 10^6` versus 17 for Tokuda. Each pass is cheap but each still touches all `N` elements, so the constant factor is poor and real-world runtime loses badly to ratio-based sequences at every practical size.

Skean et al. quantify how badly — at `N = 10^4`, Pratt-23 needs **604 502 comparisons vs Tokuda's 192 574 (3.1×)**, and at `N = 1000` it is 6.35 ms vs 3.06 ms (**2.1× slower**). The `Θ(N log²N)` bound never pays off at any size anyone sorts.

**But Pratt wins decisively on exchanges**, which is the one thing this survey would otherwise miss: at `N = 10^4` Pratt-23 does **66 923 exchanges vs Tokuda's 98 071**, and Pratt-25 does 62 191 — the best of every sequence tested, at every size. Its many cheap passes leave fewer inversions for the expensive tail. So Pratt is the right choice when writes dominate reads (memory-constrained systems, wear-sensitive storage, or comparisons that are trivially cheap) — the exact inverse of the comparison-optimal case that Ciura and Tokuda target.

Variants: Pratt-25 (`2^p·5^q`) and Pratt-34 (`3^p·4^q`) both beat Pratt-23 on comparisons (450 131 and 355 382 at `N = 10^4`) while staying near it on exchanges, so plain 3-smooth is not even the best of its own family in practice.

In this repo it is also the worst structural fit (§6): ~125 barriers instead of ~17. Keep as a theory anchor and an exchange-minimizing special case, not a default.

### Knuth 1973 — `h_{k+1} = 3h_k + 1`, i.e. `(3^k − 1)/2`, capped at `⌈N/3⌉`

Gaps `1, 4, 13, 40, 121, 364, 1093, ...`. Worst case **`Θ(N^{3/2})`**. Ratio 3, so `t(N) = log₃N ≈ 13` at `N = 10^6` — fewest passes of any practical sequence, but ratio 3 is measurably above the empirical optimum: the first pass leaves subarrays long enough that the later passes inherit more inversions.

Virtues: one-line recurrence, no floating point, no table, and the `h ≥ N/3` stopping rule is standard. The simplest defensible fallback, and the sequence to pick if the implementation must avoid `math.Pow` entirely.

### Incerpi & Sedgewick 1985 — constructed, pairwise-coprime triangle

Gaps `1, 3, 7, 21, 48, 112, ...`. Worst case **`O(N^{1 + √(8 ln(5/2) / ln N)})`** — asymptotically better than every `N^{3/2}` and `N^{4/3}` sequence, with the exponent tending to 1 as `N → ∞`. The convergence is so slow that the crossover is beyond any realistic input size, and there is no simple closed form for `h_k`. Theoretical milestone; not a candidate.

### Gonnet & Baeza-Yates 1991 — `h_k = max(⌊(5·h_{k-1} − 1)/11⌋, 1)`, `h_0 = N`

Top-down (starts at `N`, divides by ~2.2 each step): for `N = 1000`, gaps descend `454, 206, 93, 42, 19, 8, 3, 1`. Ratio 2.2, close to the empirical optimum, and integer arithmetic only.

Two caveats: worst case is **`Θ(N²)` for certain values of `N`** — the top-down division makes gap divisibility depend on `N`, so bad `N` produce bad sequences — and the published average-case figure is `0.41 N ln N (ln ln N + 1/6)` **element moves** (not comparisons), so it is not directly comparable to the comparison counts quoted for Ciura/Tokuda. Good practical performance, but the `N²` cases make it hard to recommend over Tokuda for a library.

### Tokuda 1992 — `h_k = ⌈h'_k⌉` where `h'_k = 2.25·h'_{k-1} + 1`, `h'_1 = 1`

Gaps `1, 4, 9, 20, 46, 103, 233, 525, 1182, 2660, 5985, ...`. Worst case listed as **`O(N^{4/3})`**; lower bound **`Ω(N^{1.26})`** (Zang 2026), so its exponent is pinned into a narrow band — the most precisely characterized of the practical sequences.

Ratio 2.25, `t(N) = log_{2.25}N ≈ 17` at `N = 10^6`, ~14 at `10^5`, ~9 at `10^3`. Closed-form and unbounded: computable for any `N` with no table and no extension rule, which is what separates it from Ciura. On average comparisons it sits ~3% behind Ciura in Ciura's own measurements (see below), and it wins on large inputs where Ciura's table runs out.

**The default recommendation** for this repo.

### Ciura 2001 — empirical tables, three of them

No formula: found by sequential analysis — direct search minimizing average comparisons on random input. **Best known average-case performance** of any published sequence; the paper's own claim is **~3% fewer comparisons than the best sequences known at the time** (i.e. Tokuda and Sedgewick). Worst case listed as `O(N^{4/3})`.

Important detail usually lost in secondary sources: Ciura published **three** sequences, optimized for different sizes, and only the first two are search results.

| Name | Terms | Provenance |
| --- | --- | --- |
| Ciura-128 | 1, 4, 9, 24, 85, 126 | searched, optimal for `N = 128` |
| Ciura-1000 | 1, 4, 10, 23, 57, 156, 409, 995 | searched, optimal for `N = 1000` |
| Ciura-Large | 1, 4, 10, 23, 57, 132, 301, 701, 1750 | **conjectured** to do better at much larger `N` |

What every implementation calls "Ciura's sequence" — including the row in §5 — is **Ciura-Large**, the conjectured one. Its first five terms coincide with Ciura-1000, then diverge. So the sequence in universal practical use was never the object of Ciura's optimality proof, and the `1750` term is his own, not a later extension. This matters when reading claims of Ciura's superiority: they rest on measurement, and the measured winner depends on which of the three you mean and at what `N` (see §5).

The ranking has been reproduced independently: a 2011 replication over `N = 1000…10000` (50 trials per size, uniform random) found Ciura's sequence minimizes comparisons among Ciura, Sedgewick, Incerpi–Sedgewick, powers of two, and Fibonacci — but that under a combined read+write cost model, Sedgewick–Incerpi ties with it. Consistent with Skean et al.: the comparison winner is not automatically the runtime winner.

Limits: the table stops at 701, so inputs beyond ~1500–2000 elements need an extension, conventionally `h_k = ⌊2.25·h_{k-1}⌋` (giving 1576, 3547, 7983, ...). The extension is **unproven** — Ciura's search never validated those terms, and its optimality claim does not transfer. Some implementations instead append 1750 by hand.

Practically: best choice for small-to-medium `N`; needs a documented extension rule to be a general-purpose default; costs a static table (8+ ints) instead of a recurrence.

### Lee 2021 — `h_k = ⌈(γ^k − 1)/(γ − 1)⌉`, `γ = 2.243609061420001`

Gaps `1, 4, 9, 20, 45, 102, 230, 516, 1158, 2599, 5831, ...` (arXiv:2112.11112). A γ-sequence tuned by search; empirically **fewer average comparisons than Tokuda**, which it was explicitly built to beat. Nearly identical to Tokuda from below (45 vs 46, 102 vs 103), so the gain is small and the mechanism is a better constant, not better structure — and Zang's `Ω(N^{1.26})` covers it, being a rational-geometric approximation.

Closed form, unbounded, drop-in replaceable with Tokuda. Reasonable to expose as an alternative; not enough evidence to displace Tokuda as default.

### Skean, Ehrenborg & Jaromczyk 2023 — parameterized functions, grid-searched

"Optimization Perspectives on Shellsort" (arXiv:2301.00316). Rather than one sequence, a *method*: define a parameterized gap-generating function, then grid-search its parameters against a chosen cost metric at a chosen array size. Their key structural insight is that the search space then depends only on grid granularity, **not on `N`** — unlike Ciura's direct sequence search, whose cost grows with `N` and which is why Ciura could only search up to `N = 1000`.

Two function templates. **Ours-B** is the one that reproduces cleanly:

```
k_B(i) = ⌊ a · b^(i/c) + d ⌋       a = 4.0816, b = 8.5714, c = 2.2449, d = 0
```

i.e. a geometric sequence with ratio `b^(1/c) = 2.604` and a leading offset, prefixed with 1: `1, 4, 10, 27, 72, 187, 488, ...`. Optimized for comparisons at `N = 10 000`. Note the ratio — **2.60, well above the 2.25 folk optimum** — which is direct evidence that the "best ratio" is metric- and size-dependent rather than universal.

**Ours-A** is a two-base form with floor functions in the exponents and an extra exponent parameter `f` regulating growth, optimized separately for comparisons at `N = 128` and 1000 and for *running time* at `N = 1000`. Terms: `1, 4, 9, 24, 85, 150...` (128-Comp), `1, 4, 10, 23, 57, 153, 400...` (1000-Comp), `1, 3, 7, 16, 33, 85, 179, 472...` (1000-Time). The exact form is not reproduced here — the text-extracted copy consulted had eq. (1) and the `f` column of its Table 2 corrupted, and the published terms do not reproduce under either plausible reading. Read the PDF directly if Ours-A is needed.

Results, from their Tables 3–6:

- **Ours-B10000-Comp beats Tokuda on comparisons at every size tested**, making it the best known *function-based* sequence for comparisons. It does not beat the Ciura sequences.
- Ours-A variants **match but never surpass** Ciura.
- **Ours-A1000-Time starts `1, 3, 7, 16, 33...`, not `1, 4, 9...`** — and ties Ciura-1000 for fastest wall-clock time while using *more* comparisons than the comparison-optimal sequences. Their conclusion, and the most transferable finding in the paper: a sequence that minimizes comparisons is not necessarily the fastest.
- Optimizing for **exchanges** produced no meaningful improvement over existing sequences — the Pratt family already dominates that metric (see above).

Method caveat worth knowing before citing their timings: all experiments were written in **Python**, with timing runs single-threaded on an 8-core Xeon W-3225. Comparison and exchange *counts* are implementation-independent and transfer directly; the *wall-clock* results are measured under interpreter overhead that likely compresses the differences a compiled implementation would show.

Two takeaways for this repo. Sequence choice is size- and metric-dependent, which is the argument for `WithGapSequence` being a public option rather than a constant. And their runtime/comparison divergence means benchmarking this implementation is not optional — the literature's comparison rankings do not settle which sequence is fastest here.

### Ratio families in general

Since almost every practical sequence is `h_k ≈ r^k`, the interesting question is `r`:

| `r` | Behaviour |
| --- | --- |
| 2 | Hibbard, Papernov–Stasevich. ~20 passes at `N = 10^6`. More passes than necessary; each cheap. |
| **2.2–2.3** | Gonnet–Baeza-Yates (2.2), Lee (2.2436), Tokuda/Ciura-extension (2.25). Empirical optimum across published searches optimizing **comparisons at `N ≤ 1000`** — converged on independently by different methods. ~17 passes at `10^6`. |
| 2.6 | Skean et al. Ours-B, grid-searched for comparisons at `N = 10 000`. Beats Tokuda on comparisons — evidence that the 2.25 optimum is size-specific, not universal, and drifts upward with `N`. |
| 3 | Knuth. ~13 passes; slightly past the optimum. |
| 4 | Sedgewick 1982 (`4^k + 3·2^{k-1} + 1`). ~10 passes; too aggressive, first pass leaves too much work. |
| φ ≈ 1.618 | Fibonacci-like. Too many passes; coprimality is fine but pass overhead dominates. |

Integer ratios that are prime powers (2, 3, 4, 8) are the danger zone: the `+1` or `−1` correction terms in Hibbard, Papernov–Stasevich, Knuth, and Sedgewick exist precisely to break divisibility between consecutive gaps. Any custom sequence must do the same — a bare `⌈r^k⌉` with integer `r` reintroduces Shell's `Θ(N²)` failure mode.

## 5. Summary table

Baselines in *italics*.

| Sequence | Formula | First terms | `r` | `t(10^6)` | Worst case | Average case | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Shell 1959 | `⌊N/2^k⌋` | N/2, N/4, … | 2 | 20 | `Θ(N²)` | poor | divisibility failure |
| Frank–Lazarus 1960 | `2⌊N/2^{k+1}⌋+1` | …, 3, 1 | 2 | 20 | `Θ(N^{3/2})` | — | odd-forced Shell |
| *Hibbard 1963* | `2^k − 1` | 1, 3, 7, 15, 31 | 2 | 20 | `Θ(N^{3/2})` | conj. `N^{5/4}` (unproven) | repo alternative |
| Papernov–Stasevich 1965 | `2^k + 1` | 1, 3, 5, 9, 17 | 2 | 20 | `Θ(N^{3/2})` | — | ≈ Hibbard |
| Pratt 1971 | `2^p·3^q` | 1, 2, 3, 4, 6, 8, 9 | — | ~125 | **`Θ(N log²N)`** | best asymptotic | worst constant factor |
| Knuth 1973 | `3h+1` | 1, 4, 13, 40, 121 | 3 | 13 | `Θ(N^{3/2})` | good | simplest good choice |
| Incerpi–Sedgewick 1985 | constructed | 1, 3, 7, 21, 48, 112 | — | — | `O(N^{1+√(8ln(5/2)/ln N)})` | — | crossover unreachable |
| Sedgewick 1982 | `4^k+3·2^{k−1}+1` | 1, 8, 23, 77, 281 | 4 | 10 | `O(N^{4/3})` | — | ratio too high |
| *Sedgewick 1986* | piecewise even/odd `k` | 1, 5, 19, 41, 109 | ~2 | ~20 | `O(N^{4/3})` | good | **repo default** |
| Gonnet–Baeza-Yates 1991 | `max(⌊(5h/11)⌋,1)` top-down | 454, 206, 93, … (N=1000) | 2.2 | 17 | `Θ(N²)` for some `N` | `0.41 N ln N(ln ln N + 1/6)` moves | `N`-dependent risk |
| **Tokuda 1992** | `⌈h'⌉`, `h' = 2.25h'+1` | 1, 4, 9, 20, 46, 103 | 2.25 | 17 | `O(N^{4/3})`, `Ω(N^{1.26})` | ~3% behind Ciura | **recommended default** |
| **Ciura 2001** | 3 empirical tables | Large: 1, 4, 10, 23, 57, 132, 301, 701, 1750 | 2.25 (ext.) | 17 | `O(N^{4/3})` | best known; ties Tokuda within ~2% | "the" Ciura sequence is his *conjectured* one |
| Lee 2021 | `⌈(γ^k−1)/(γ−1)⌉`, γ=2.2436 | 1, 4, 9, 20, 45, 102 | 2.2436 | 17 | `Ω(N^{1.26})` | beats Tokuda (measured) | closed form, unbounded |
| Skean et al. 2023 (Ours-B) | `⌊a·b^{i/c} + d⌋` | 1, 4, 10, 27, 72, 187, 488 | **2.604** | 15 | — | beats Tokuda on comparisons; 2× its exchanges | grid-searched per size + metric |

Pass counts are `log_r 10^6` (Pratt: `(log₂N)(log₃N)/2`), rounded — asymptotic estimates, not exact enumerations.

**On the average-case column.** Worst-case entries are theorems; average-case entries are measurements. Ciura's own per-size tables sit behind the FCT 2001 paywall with no open mirror, so his `~3%` is a summary claim rather than reproduced data — but Skean et al. 2023 independently re-measured the whole field, and those numbers are in §5.1 below.

## 5.1 Measured operation counts

From Skean et al. 2023, Tables 3–5: mean over 1000 random permutations of `1..N`, Fisher–Yates shuffled. `μCO` = comparisons, `μEX` = exchanges. Ciura sequences extended past their last term by a 2.25 geometric ratio where `N` required it.

| Sequence | N=200 μCO | N=200 μEX | N=1000 μCO | N=1000 μEX | N=10000 μCO | N=10000 μEX |
| --- | --- | --- | --- | --- | --- | --- |
| Ciura-128 | 1800 | 970 | 13 300 | 7003 | 195 256 | 105 544 |
| Ciura-1000 | 1787 | 920 | 12 918 | 7002 | 193 778 | 111 338 |
| **Ciura-Large** | 1794 | 907 | 13 035 | 6701 | **191 435** | 101 680 |
| **Tokuda** | 1808 | 891 | 13 116 | 6556 | 192 574 | **98 071** |
| Ours-B (Skean) | 1775 | 960 | 12 980 | 7245 | 192 029 | 209 292 |
| Pratt-23 | 4095 | 589 | 34 380 | 4253 | 604 502 | 66 923 |
| Pratt-25 | 3207 | 610 | 26 211 | 4318 | 450 131 | 62 191 |
| Pratt-34 | 2593 | 660 | 20 974 | 4671 | 355 382 | 63 272 |

Wall-clock, `N = 1000`, single-threaded Python: Ciura-1000 and Ours-A1000-Time both 3.01 ms, Ciura-Large 3.04, Tokuda 3.06, Ciura-128 3.07, Pratt-34 4.17, Pratt-25 5.00, Pratt-23 6.35.

### Reproduced in this repo

**What this section is for.** The gap sequences in `lib/gaps.go` were written from formulas in papers. A formula can be transcribed slightly wrong and still produce a sequence that looks entirely reasonable — ascending, starting at 1, growing at about the right rate — and still be the wrong sequence. Ordinary tests cannot catch that, because there is nothing to compare against except the formula that was already transcribed.

So instead: run our sorting code, count the operations it performs, and check those counts against counts published in a paper. If our generated gaps differ from the paper's, the counts come out different. If they match, the gaps match. It is an indirect check with a very sharp edge — comparison counts run into the hundreds of thousands, and agreeing on one to four decimal places by accident is not plausible.

**How it was run.** `TestOperationCountsMatchLiterature` (`lib/gaps_bench_test.go`) sorts 200 randomly shuffled arrays of 10 000 elements with each sequence, counts comparisons and exchanges (defined in §1) on every one, and averages. The averages are compared against Skean et al.'s published Tables 3–5. "Δ" is how far our average lands from theirs, in percent; anything under a couple of percent is agreement.

| Sequence | μCO measured | μCO published | Δ | μEX measured | μEX published | Δ |
| --- | --- | --- | --- | --- | --- | --- |
| Tokuda | 192 482 | 192 574 | −0.048% | 98 176 | 98 071 | +0.107% |
| Ciura-Large | 191 382 | 191 435 | −0.028% | 101 840 | 101 680 | +0.158% |
| Pratt-23 | 604 438 | 604 502 | −0.011% | **63 705** | **66 923** | **−4.81%** |
| Pratt-25 | 449 982 | 450 131 | −0.033% | 62 357 | 62 191 | +0.267% |

Seven of the eight numbers land within 0.27% of the published ones. That is a strong result: the gap generators, the operation counter, and our reading of Skean et al.'s definitions are all confirmed correct in one shot.

#### The eighth number, explained

One cell disagrees: Pratt-23's exchange count. We measure 63 705 where the paper prints 66 923 — ours is 4.81% lower.

The obvious suspect is our own code, and that was the working assumption. Three checks say otherwise:

1. **The same sequence's comparison count is right.** Pratt-23's comparisons match to 0.011% — 604 438 against a published 604 502, a difference of 64 out of 604 502. Comparisons depend on *every* gap the sequence produces (67 of them at this size), so if our list of gaps were wrong that number could not land that close. The gaps are right.

2. **No other gap list explains their pair of numbers.** If the paper had built its Pratt gaps with a different cutoff, *both* of its numbers would shift together. That was tested directly: stopping the sequence at `N/2`, `N/3` or `N/4` does push exchanges up toward 66 923 — but it drags comparisons down to 580 166, 553 377 and 534 384, destroying the match with their 604 502. No cutoff reproduces both published numbers at once, meaning their two numbers cannot both come from one gap sequence.

3. **The neighbouring sequence reproduces perfectly.** Pratt-25 runs through the identical code with one constant changed (5 instead of 3), and both its numbers match (+0.033% and +0.267%). Same counter, same generator — so neither has a systematic fault.

Conclusion: the disagreement is in that one published cell, not in this catalog. The test still asserts a value there — our measured 63 705, at the same tolerance — so if the Pratt generator is ever changed in a way that moves this number, the test fails and says so. It simply stops pretending the published number is the one to match.

**Effect on the conclusions below:** point 2, the comparison/exchange tradeoff, **holds and gets slightly stronger** — Pratt-23 uses 65% of Tokuda's exchanges rather than 68%. Point 3 concerns wall-clock time and is unaffected.

Four things this table settles:

1. **Ciura-vs-Tokuda is a coin flip, not a 3% win.** At `N = 10 000` the spread across all four Ciura/Tokuda variants is 191 435–195 256 comparisons — **under 2%**, with Tokuda ahead of two of the three Ciura variants. Ciura's `~3%` advantage was measured against the state of the art of 2001 at `N ≤ 1000`; it does not survive as a general claim.
2. **The comparison/exchange tradeoff is large and opposite.** Pratt-23 uses 3.1× Tokuda's comparisons but only 68% of its exchanges. Choosing by comparisons alone silently picks the write-heavy option.
3. **Pratt's cost is real and grows.** 6.35 ms vs 3.06 ms at `N = 1000` — no crossover in sight, consistent with §3's constant-factor warning.
4. **Ours-B's exchange count is an outlier** (209 292 at `N = 10 000`, 2× Tokuda) — it was grid-searched for comparisons only, and shows what single-metric optimization costs on the metric you ignored.

Practical consequence for this repo: **the top group (Ciura variants, Tokuda, Lee, Ours-B) is separated by ~2% on comparisons**, well inside what the insertion-sort inner loop, goroutine spawn cost, and cache behaviour will swamp here. Choose within that group by benchmarking the actual implementation, not from this table. The table's real use is the two large, robust effects: that group vs the Pratt family, and comparisons vs exchanges.

## 6. Fit with this repo's parallel design

`ShellSort` runs one goroutine per subarray with a `sync.WaitGroup` barrier between passes. This section assumes, per the design question being asked, **enough CPUs for every subarray** — unbounded parallel width, no core-count ceiling. It deliberately ignores the current `< 5` threshold (`lib/sort.go:19`) and the `d*3 < length` guard (`lib/sort.go:41`); both are implementation questions, not properties of the sequences.

Under that model, per pass with gap `h`:

- **parallel width** = `h` (number of independent subarrays);
- **work** = `h · cost(N/h)` ≈ total elements touched;
- **span** (critical path) = `cost(N/h)`, the longest single subarray;
- **one barrier** per pass.

Three consequences, in order of importance:

**1. The `h = 1` pass is an unremovable serial floor.** It is one subarray of `N` elements, width 1, span `Ω(N)`. Every sequence ends with it. So total span `≈ Σ_k cost(N/h_k)`, a geometric-ish sum **dominated by the last few passes** — which are exactly the passes with the least parallel width. The passes that parallelize beautifully (large `h`, many short subarrays) are the cheap ones; the expensive tail passes barely parallelize at all. Amdahl caps the achievable speedup regardless of gap sequence, and **no choice in §4 changes this** — it is a property of Shellsort's structure, not of the gaps.

**2. Pass count is barrier count, so ratio ~2.25 wins twice.** Every pass costs a `WaitGroup` round-trip plus goroutine setup for `h` goroutines. Tokuda/Ciura/Lee need ~17 barriers at `N = 10^6`; Hibbard, Papernov–Stasevich, and the current Sedgewick 1986 need ~20; **Pratt needs ~125**. Pratt's `Θ(log²N)` structure buys `O(N)` work per pass but adds no span improvement — the same `Ω(N)` tail — while multiplying synchronization by 7×. Pratt has the best worst-case bound and the worst fit for this implementation. Knuth's 13 barriers are the fewest, at the cost of a slightly suboptimal ratio.

**3. Fixed tables cap parallel width.** Ciura's largest gap is 701, so without an extension the first pass has width 701 and subarrays of length `N/701` — at `N = 10^6` that is ~1400 elements insertion-sorted serially inside each goroutine, in the very first pass, which is where the sequential cost concentrates. Closed-form sequences (Tokuda, Lee, Knuth) scale their first gap with `N` and keep the first pass's subarrays short. **For a parallel implementation this argues more strongly for Tokuda over Ciura than the sequential literature does**, since Ciura's advantage is in average comparisons while its fixed ceiling directly costs parallel width.

Secondary notes:

- **Goroutine count per pass = `h`.** With a closed-form sequence the first gap is `Θ(N)`-ish, meaning thousands of goroutines in the first pass, each doing very little. Spawning cost then dominates real work even though the model says width is free — a chunking strategy (one goroutine per group of subarrays) decouples parallel width from goroutine count, and is independent of which sequence is chosen.
- **Locality inverts under parallelism.** Large-gap passes stride badly in the sequential model, but split across goroutines each subarray is a strided walk over the whole array — several goroutines fighting for the same cache lines. Small-gap passes are cache-friendly *and* poorly parallel. So cache behaviour and parallel width trade off in the same direction as span: the tail passes are where the time goes.
- **Gap-table cost is negligible** in every case (`O(log_r N)` ints, ~17 for Tokuda at `N = 10^6`), but note that a closed-form sequence can be generated with `append` and no `N`-dependence, while Gonnet–Baeza-Yates must be generated top-down from `N` and Ciura must be stored.

## 7. Recommendation

For `WithGapSequence` (`docs/MODERNIZATION_PLAN.md` §A):

1. **Tokuda 1992 as the default.** Closed form, unbounded, ~17 passes at `10^6`, worst case boxed in between `Ω(N^{1.26})` and `O(N^{4/3})`, and — per §5.1 — statistically tied with every Ciura variant on comparisons at `N = 10^4` while beating all of them on **exchanges** (98 071 vs 101 680–111 338). It also has the largest first gap of the strong candidates, which matters most for the parallel structure. The measured data strengthens this choice rather than merely permitting it.
2. **Ciura-Large as an opt-in** for inputs in the low thousands, where his search actually applies, with two caveats documented: the sequence in common use is his *conjectured* variant, and the `⌊2.25h⌋` extension past 1750 is unvalidated.
3. **Knuth 1973 as the zero-dependency fallback** — integer arithmetic, one-line recurrence, no `math.Pow`, `Θ(N^{3/2})` proven.
4. **Lee 2021 as an experimental option**, since it is a two-line change from Tokuda and reportedly beats it on average comparisons.
5. **Pratt as documentation only — with one exception.** Best worst-case bound in the literature, worst practical constant (3.1× Tokuda's comparisons at `N = 10^4`), ~125 barriers here. But it is the **exchange-minimizing** family by a wide margin (§5.1), so if a consumer's cost model is dominated by writes rather than comparisons, Pratt-25 (`2^p·5^q`) is the right answer and worth exposing for that case alone.
6. **Do not add** Shell, Frank–Lazarus, Papernov–Stasevich (dominated), or Gonnet–Baeza-Yates (`Θ(N²)` for some `N`).

Retain Sedgewick 1986 and Hibbard for continuity, as `MODERNIZATION_PLAN.md` already assumes.

Open question worth measuring rather than reasoning about, per Skean et al.: whether the comparison-optimal sequence is also the runtime-optimal one under this parallel scheme. It probably is not, and the answer is size-dependent — which is the case for keeping the sequence a public option instead of a constant.

## 8. References

- Shell, D. L. (1959). "A High-Speed Sorting Procedure." *CACM* 2(7).
- Frank, R. M. & Lazarus, R. B. (1960). "A High-Speed Sorting Procedure." *CACM* 3(1).
- Hibbard, T. N. (1963). "An Empirical Study of Minimal Storage Sorting." *CACM* 6(5).
- Papernov, A. A. & Stasevich, G. V. (1965). "A Method of Information Sorting in Computer Memories." *Problems of Information Transmission* 1(3).
- Pratt, V. R. (1971/1972). *Shellsort and Sorting Networks.* PhD thesis, Stanford.
- Knuth, D. E. (1973). *The Art of Computer Programming*, Vol. 3, §5.2.1.
- Incerpi, J. & Sedgewick, R. (1985). "Improved Upper Bounds on Shellsort." *JCSS* 31(2).
- Sedgewick, R. (1986). "A New Upper Bound for Shellsort." *Journal of Algorithms* 7(2).
- Gonnet, G. H. & Baeza-Yates, R. (1991). *Handbook of Algorithms and Data Structures*, 2nd ed.
- Tokuda, N. (1992). "An Improved Shellsort." *IFIP Congress*.
- Plaxton, C. G., Poonen, B. & Suel, T. (1992). "Improved Lower Bounds for Shellsort." *FOCS*.
- Jiang, T., Li, M. & Vitányi, P. (2000). "Average-Case Complexity of Shellsort." *JACM* 47(5).
- Ciura, M. (2001). "Best Increments for the Average Case of Shellsort." *FCT 2001*, LNCS 2138, pp. 106–117. [doi:10.1007/3-540-44669-9_12](https://doi.org/10.1007/3-540-44669-9_12) — paywalled; no open mirror located.
- Pigeon, S. (2011). "Shellsort." *Harder, Better, Faster, Stronger* — independent replication over `N = 1000…10000`, 50 trials/size; open-access measurement of the comparison ranking. <https://hbfs.wordpress.com/2011/03/01/shellsort/>
- Janson, S. & Knuth, D. E. (1997). "Shellsort with Three Increments." *Random Structures & Algorithms* 10.
- Lee, Y. (2021). "Empirically Improved Tokuda Gap Sequence in Shellsort." arXiv:2112.11112.
- Skean, O., Ehrenborg, R. & Jaromczyk, J. W. (2023). "Optimization Perspectives on Shellsort." arXiv:2301.00316. Source of all measured counts in §5.1. Eq. (1) and the `f` column of Table 2 were unrecoverable from the text extraction consulted, so the Ours-A closed form is described but not reproduced.
- Sedgewick, R. (1996). "Analysis of Shellsort and Related Algorithms." *ESA 1996* — survey; the standard full history of gap sequences.
- Selmer, E. S. (1987). "On Shellsort and the Frobenius Problem" — the Frobenius-number technique underlying the `O(N)`-per-pass arguments in §1.
- Zang, Z. (2026). "Improved lower bounds of the time complexity of shellsort." arXiv:2607.08997 (preprint, unrefereed).
- Wikipedia, "Shellsort" — used to cross-check formulae and worst-case cells against the primary sources above.
