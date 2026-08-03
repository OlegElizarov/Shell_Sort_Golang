package lib

import (
	"math"
	"slices"
)

// GapSequence pairs a gap-sequence generator with the name used to label it in
// test failures and benchmark output.
//
// Every Gaps implementation in this package obeys the same contract, which
// TestGapSequenceContract enforces for the whole catalog:
//
//   - the result is ascending and strictly increasing;
//   - the first element is 1;
//   - every element is < length;
//   - the result is never empty — a length below 2 still yields []int{1}.
//
// The length parameter is not incidental. Shell, FrankLazarus and
// GonnetBaezaYates derive their gaps from the input size rather than
// generating bottom-up, and the Ciura tables need it to decide where to start
// extending.
// Field order is Gaps before Name to satisfy govet's fieldalignment check.
type GapSequence struct {
	// Gaps returns the gaps to use for an input of the given length,
	// ascending. Callers that need descending order iterate in reverse.
	Gaps func(length int) []int
	// Name labels the sequence in test and benchmark output.
	Name string
}

// Tokuda is h_k = ⌈h'_k⌉ where h'_k = 2.25·h'_{k-1} + 1 and h'_1 = 1:
// 1, 4, 9, 20, 46, 103, 233, 525. Closed form, unbounded, ratio 2.25, ~17
// passes at N = 10^6. The default: it ties every Ciura variant on comparisons
// at N = 10^4 while using the fewest exchanges of the practical sequences, and
// its first gap scales with N, which matters for the parallel structure here.
// Requires floating point.
var Tokuda = GapSequence{Name: "Tokuda", Gaps: tokudaGaps}

// Ciura is the sequence in universal practical use — Ciura's 2001 table
// 1, 4, 10, 23, 57, 132, 301, 701, 1750 — extended past its last term by the
// ⌊2.25h⌋ rule. Best measured average comparisons.
//
// Two caveats worth knowing before trusting it as a default. This table is the
// one Ciura *conjectured* would do better at large N; only Ciura128 and
// Ciura1000 are search results. And the ⌊2.25h⌋ extension is the convention
// adopted by Skean et al., not something Ciura validated — inputs past ~1750
// are running on an unvalidated tail.
var Ciura = GapSequence{Name: "Ciura", Gaps: func(length int) []int {
	return extendedTableGaps(ciuraLargeTable, length)
}}

// Ciura128 is Ciura's table searched for N = 128: 1, 4, 9, 24, 85, 126.
// Optimal for that size, a small-N reference elsewhere. Extended by ⌊2.25h⌋
// past its last term, which carries none of the search's optimality.
var Ciura128 = GapSequence{Name: "Ciura128", Gaps: func(length int) []int {
	return extendedTableGaps(ciura128Table, length)
}}

// Ciura1000 is Ciura's table searched for N = 1000:
// 1, 4, 10, 23, 57, 156, 409, 995. Optimal for that size, a mid-N reference
// elsewhere. Extended by ⌊2.25h⌋ past its last term, which carries none of the
// search's optimality.
var Ciura1000 = GapSequence{Name: "Ciura1000", Gaps: func(length int) []int {
	return extendedTableGaps(ciura1000Table, length)
}}

// Knuth is h_{k+1} = 3h_k + 1 stopped at ⌈N/3⌉: 1, 4, 13, 40, 121, 364.
// Worst case Θ(N^{3/2}), ratio 3, ~13 passes at N = 10^6 — the fewest of any
// practical sequence, at a ratio measurably above the empirical optimum. The
// zero-dependency fallback: one-line recurrence, integer arithmetic only.
var Knuth = GapSequence{Name: "Knuth", Gaps: knuthGaps}

// Lee is h_k = ⌈(γ^k − 1)/(γ − 1)⌉ with γ = 2.243609061420001:
// 1, 4, 9, 20, 45, 102, 230, 516 (arXiv:2112.11112). A Tokuda variant tuned by
// search, sitting just below Tokuda at every term; reportedly fewer average
// comparisons. Closed form, unbounded, requires floating point.
var Lee = GapSequence{Name: "Lee", Gaps: leeGaps}

// SkeanB is ⌊a·b^(i/c) + d⌋ with a = 4.0816, b = 8.5714, c = 2.2449, d = 0,
// prefixed with 1: 1, 4, 10, 27, 72, 187, 488 (arXiv:2301.00316, "Ours-B").
// Grid-searched for comparisons at N = 10 000, which is why its ratio is 2.604
// rather than the 2.25 folk optimum — evidence that the best ratio is
// size- and metric-dependent. Beats Tokuda on comparisons and uses roughly
// twice its exchanges. Requires floating point.
var SkeanB = GapSequence{Name: "SkeanB", Gaps: skeanBGaps}

// Pratt is every 3-smooth number 2^p·3^q below N:
// 1, 2, 3, 4, 6, 8, 9, 12, 16. Worst case Θ(N log²N) — the best proven bound
// of any sequence — and the worst practical constant: 3.1× Tokuda's
// comparisons at N = 10^4, and Θ(log²N) passes (~125 at N = 10^6) means ~7×
// the synchronization barriers here.
//
// It wins decisively on exchanges, so it is the right answer when writes
// dominate reads. Integer arithmetic only.
var Pratt = GapSequence{Name: "Pratt", Gaps: func(length int) []int {
	return smoothGaps(length, 2, 3)
}}

// Pratt25 is every 2^p·5^q below N: 1, 2, 4, 5, 8, 10, 16, 20. Fewest
// exchanges of any sequence measured, at every size, and fewer comparisons
// than plain Pratt. Integer arithmetic only.
var Pratt25 = GapSequence{Name: "Pratt25", Gaps: func(length int) []int {
	return smoothGaps(length, 2, 5)
}}

// Pratt34 is every 3^p·4^q below N: 1, 3, 4, 9, 12, 16, 27. The
// comparison-cheapest of the Pratt family while staying near it on exchanges.
// Integer arithmetic only.
var Pratt34 = GapSequence{Name: "Pratt34", Gaps: func(length int) []int {
	return smoothGaps(length, 3, 4)
}}

// Sedgewick86 is the piecewise 9·2^k − 9·2^{k/2} + 1 for even k and
// 8·2^k − 6·2^{(k+1)/2} + 1 for odd k: 1, 5, 19, 41, 109. Worst case
// O(N^{4/3}). This package's historical default, kept for continuity.
// Integer arithmetic only.
var Sedgewick86 = GapSequence{Name: "Sedgewick86", Gaps: sedgewick86Gaps}

// Sedgewick82 is 4^k + 3·2^{k−1} + 1 prefixed with 1: 1, 8, 23, 77, 281.
// Worst case O(N^{4/3}). Ratio 4 is too aggressive — the first pass leaves too
// much work for the rest — so it serves as the high-ratio datapoint in the
// study. Integer arithmetic only.
var Sedgewick82 = GapSequence{Name: "Sedgewick82", Gaps: sedgewick82Gaps}

// Hibbard is 2^k − 1: 1, 3, 7, 15, 31. Worst case Θ(N^{3/2}), ratio 2, so
// ~20 passes at N = 10^6 — more than a ratio-2.25 sequence needs, for no gain.
// Kept as a selectable alternative. Integer arithmetic only.
var Hibbard = GapSequence{Name: "Hibbard", Gaps: hibbardGaps}

// PapernovStasevich is 2^k + 1 prefixed with 1: 1, 3, 5, 9, 17, 33. Worst case
// Θ(N^{3/2}), matching Hibbard with a slightly better constant in practice.
// Superseded by Knuth; kept as the ratio-2 control. Integer arithmetic only.
var PapernovStasevich = GapSequence{Name: "PapernovStasevich", Gaps: papernovStasevichGaps}

// IncerpiSedgewick is the constructed pairwise-coprime sequence
// 1, 3, 7, 21, 48, 112, 336, 861, 1968. Worst case
// O(N^{1+√(8ln(5/2)/ln N)}) — asymptotically better than every N^{3/2} and
// N^{4/3} sequence, with the crossover beyond any realistic input size. There
// is no closed form for h_k, so this is a table (OEIS A036569) rather than a
// recurrence, and it is not extended past its last tabulated term.
var IncerpiSedgewick = GapSequence{Name: "IncerpiSedgewick", Gaps: func(length int) []int {
	return tableGaps(incerpiSedgewickTable, length)
}}

// FrankLazarus is 2⌊N/2^{k+1}⌋ + 1, descending to 1 — Shell's halving forced
// odd, which breaks the shared factor between consecutive gaps and drops the
// worst case from Θ(N²) to Θ(N^{3/2}). The cheapest fix in the algorithm's
// history and the clearest evidence that coprimality, not growth rate, is what
// matters. N-derived control. Integer arithmetic only.
var FrankLazarus = GapSequence{Name: "FrankLazarus", Gaps: frankLazarusGaps}

// GonnetBaezaYates is h_k = max(⌊5·h_{k-1}/11⌋, 1) from h_0 = N, i.e.
// top-down division by ~2.2: for N = 1000 it descends 454, 206, 93, 42, 19, 8,
// 3, 1. Good practical ratio and integer arithmetic only, but the worst case
// is Θ(N²) for certain N — the top-down division makes gap divisibility depend
// on N — so it is a control here, not a recommendation.
//
// Sources disagree on whether the numerator is 5h or 5h−1; the two differ only
// when 11 divides 5h. This implements ⌊5h/11⌋, matching the Handbook of
// Algorithms pseudo-code.
var GonnetBaezaYates = GapSequence{Name: "GonnetBaezaYates", Gaps: gonnetBaezaYatesGaps}

// Shell is the original ⌊N/2^k⌋: N/2, N/4, …, 1. Worst case Θ(N²) — no better
// than insertion sort, because every gap divides the previous one. The
// negative control: it shows that ratio alone does not determine quality.
// Integer arithmetic only.
var Shell = GapSequence{Name: "Shell", Gaps: shellGaps}

// Gap tables that have no closed form and must be stored verbatim.
var (
	// ciuraLargeTable is Ciura's conjectured large-N sequence.
	ciuraLargeTable = []int{1, 4, 10, 23, 57, 132, 301, 701, 1750}
	// ciura128Table is Ciura's search result for N = 128.
	ciura128Table = []int{1, 4, 9, 24, 85, 126}
	// ciura1000Table is Ciura's search result for N = 1000.
	ciura1000Table = []int{1, 4, 10, 23, 57, 156, 409, 995}
	// incerpiSedgewickTable is OEIS A036569. Terms past 112 are transcribed
	// from the published table, not re-derived from the construction.
	incerpiSedgewickTable = []int{
		1, 3, 7, 21, 48, 112, 336, 861, 1968, 4592, 13776, 33936,
		86961, 198768, 463792, 1391376, 3402672, 8382192, 21479367,
		49095696, 114556624, 343669872, 852913488, 2085837936,
	}
)

// tokudaRatio is the growth factor of Tokuda's recurrence, and by convention
// also the extension rule for the Ciura tables.
const tokudaRatio = 2.25

// leeGamma is Lee's searched growth factor (arXiv:2112.11112).
const leeGamma = 2.243609061420001

// Skean et al.'s "Ours-B" parameters, grid-searched for comparisons at
// N = 10 000 (arXiv:2301.00316). Their d is 0 and is folded away here.
const (
	skeanFactor   = 4.0816
	skeanBase     = 8.5714
	skeanExponent = 2.2449
)

func tokudaGaps(length int) []int {
	if length < 2 {
		return []int{1}
	}
	var gaps []int
	for h := 1.0; ; h = tokudaRatio*h + 1 {
		g := int(math.Ceil(h))
		if g >= length {
			return gaps
		}
		gaps = append(gaps, g)
	}
}

func knuthGaps(length int) []int {
	if length < 2 {
		return []int{1}
	}
	// Knuth stops the generator at ⌈N/3⌉ rather than filtering the gaps, so
	// the cap belongs here and not on the shared sorting path.
	limit := (length + 2) / 3
	gaps := []int{1}
	for h := 4; h < length && h < limit; h = 3*h + 1 {
		gaps = append(gaps, h)
	}
	return gaps
}

func leeGaps(length int) []int {
	if length < 2 {
		return []int{1}
	}
	var gaps []int
	for pow := leeGamma; ; pow *= leeGamma {
		g := int(math.Ceil((pow - 1) / (leeGamma - 1)))
		if g >= length {
			return gaps
		}
		gaps = append(gaps, g)
	}
}

func skeanBGaps(length int) []int {
	if length < 2 {
		return []int{1}
	}
	// The parameterized function starts at 4; the leading 1 is prefixed, as
	// in the paper.
	gaps := []int{1}
	for i := 0; ; i++ {
		g := int(skeanFactor * math.Pow(skeanBase, float64(i)/skeanExponent))
		if g >= length {
			return gaps
		}
		gaps = append(gaps, g)
	}
}

// smoothGaps returns every {a,b}-smooth number below length, ascending — the
// Pratt family. a and b must be coprime, or the two progressions collide and
// the result contains duplicates. This is the only family whose gap count is
// Θ(log²N) rather than Θ(log N).
func smoothGaps(length, a, b int) []int {
	if length < 2 {
		return []int{1}
	}
	var gaps []int
	for x := 1; x < length; x *= a {
		for y := x; y < length; y *= b {
			gaps = append(gaps, y)
		}
	}
	slices.Sort(gaps)
	return gaps
}

func sedgewick86Gaps(length int) []int {
	if length < 2 {
		return []int{1}
	}
	var gaps []int
	for k := 0; ; k++ {
		var g int
		if k%2 == 0 {
			g = 9*(1<<k) - 9*(1<<(k/2)) + 1
		} else {
			g = 8*(1<<k) - 6*(1<<((k+1)/2)) + 1
		}
		if g >= length {
			return gaps
		}
		gaps = append(gaps, g)
	}
}

func sedgewick82Gaps(length int) []int {
	if length < 2 {
		return []int{1}
	}
	gaps := []int{1}
	for k := 1; ; k++ {
		g := 1<<(2*k) + 3*(1<<(k-1)) + 1
		if g >= length {
			return gaps
		}
		gaps = append(gaps, g)
	}
}

func hibbardGaps(length int) []int {
	if length < 2 {
		return []int{1}
	}
	var gaps []int
	for h := 1; h < length; h = 2*h + 1 {
		gaps = append(gaps, h)
	}
	return gaps
}

func papernovStasevichGaps(length int) []int {
	if length < 2 {
		return []int{1}
	}
	gaps := []int{1}
	for k := 1; ; k++ {
		g := 1<<k + 1
		if g >= length {
			return gaps
		}
		gaps = append(gaps, g)
	}
}

func frankLazarusGaps(length int) []int {
	if length < 2 {
		return []int{1}
	}
	// Derived top-down from N, so build descending and reverse. The first
	// term 2⌊N/4⌋+1 is always below N for N ≥ 2.
	var desc []int
	for k := 1; ; k++ {
		g := 2*(length>>(k+1)) + 1
		desc = append(desc, g)
		if g == 1 {
			break
		}
	}
	slices.Reverse(desc)
	return desc
}

func gonnetBaezaYatesGaps(length int) []int {
	if length < 2 {
		return []int{1}
	}
	var desc []int
	for h := length; h > 1; {
		h = max(5*h/11, 1)
		desc = append(desc, h)
	}
	slices.Reverse(desc)
	return desc
}

func shellGaps(length int) []int {
	if length < 2 {
		return []int{1}
	}
	var desc []int
	for h := length / 2; h > 0; h /= 2 {
		desc = append(desc, h)
	}
	slices.Reverse(desc)
	return desc
}

// tableGaps returns the tabulated gaps below length, without extension. The
// returned slice is a fresh copy, so callers cannot mutate the table.
func tableGaps(table []int, length int) []int {
	if length < 2 {
		return []int{1}
	}
	var gaps []int
	for _, g := range table {
		if g >= length {
			break
		}
		gaps = append(gaps, g)
	}
	return gaps
}

// extendedTableGaps returns the tabulated gaps below length, continuing past
// the table by the ⌊2.25h⌋ rule when the input outgrows it. The extension is
// convention, not a search result — see the Ciura doc comment.
func extendedTableGaps(table []int, length int) []int {
	gaps := tableGaps(table, length)
	if length < 2 || len(gaps) < len(table) {
		return gaps
	}
	for h := gaps[len(gaps)-1]; ; {
		// ⌊2.25h⌋ in integer arithmetic.
		h = h * 9 / 4
		if h >= length {
			return gaps
		}
		gaps = append(gaps, h)
	}
}
