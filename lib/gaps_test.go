package lib

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

// allSequences is the whole catalog. Tests and benchmarks range over it rather
// than naming entries one at a time, so a new GapSequence is covered by the
// contract, correctness and benchmark suites the moment it is added here.
var allSequences = []GapSequence{
	Tokuda,
	Ciura,
	Ciura128,
	Ciura1000,
	Knuth,
	Lee,
	SkeanB,
	Pratt,
	Pratt25,
	Pratt34,
	Sedgewick86,
	Sedgewick82,
	Hibbard,
	PapernovStasevich,
	IncerpiSedgewick,
	FrankLazarus,
	GonnetBaezaYates,
	Shell,
}

// contractLengths spans the degenerate inputs (0, 1, 2), the sizes Ciura
// searched (128, 1000), and one large enough to exhaust every fixed table.
var contractLengths = []int{0, 1, 2, 16, 128, 1000, 50000}

func TestGapSequenceContract(t *testing.T) {
	t.Parallel()
	for _, seq := range allSequences {
		t.Run(seq.Name, func(t *testing.T) {
			t.Parallel()
			for _, length := range contractLengths {
				t.Run(strconv.Itoa(length), func(t *testing.T) {
					t.Parallel()
					gaps := seq.Gaps(length)

					require.NotEmpty(t, gaps, "gaps must never be empty")
					require.Equal(t, 1, gaps[0], "first gap must be 1")

					if length < 2 {
						require.Equal(t, []int{1}, gaps,
							"a length below 2 must yield exactly []int{1}")
						return
					}

					for i := 1; i < len(gaps); i++ {
						require.Greater(t, gaps[i], gaps[i-1],
							"gaps must be strictly increasing: %v", gaps)
					}
					require.Less(t, gaps[len(gaps)-1], length,
						"every gap must be below length: %v", gaps)
				})
			}
		})
	}
}

// TestGapSequenceGoldenTerms pins the leading terms of every sequence against
// the values transcribed from the primary sources in docs/GAP_SEQUENCES.md.
// The contract test above only proves a sequence is well-formed; this is what
// catches a plausible-looking formula that generates the wrong sequence.
func TestGapSequenceGoldenTerms(t *testing.T) {
	t.Parallel()
	cases := []struct {
		seq GapSequence
		// want is the expected leading terms at the given length.
		want []int
		// length is the input the terms were transcribed for.
		length int
		// exact requires want to be the whole sequence, not just a prefix.
		// Set for the sequences derived from N, where the tail is part of
		// the claim being tested.
		exact bool
	}{
		{seq: Tokuda, length: 50000, want: []int{1, 4, 9, 20, 46, 103, 233, 525}},
		{seq: Ciura, length: 50000, want: []int{1, 4, 10, 23, 57, 132, 301, 701, 1750}},
		{seq: Ciura128, length: 50000, want: []int{1, 4, 9, 24, 85, 126}},
		{seq: Ciura1000, length: 50000, want: []int{1, 4, 10, 23, 57, 156, 409, 995}},
		{seq: Knuth, length: 50000, want: []int{1, 4, 13, 40, 121, 364}},
		{seq: Lee, length: 50000, want: []int{1, 4, 9, 20, 45, 102, 230, 516}},
		{seq: SkeanB, length: 50000, want: []int{1, 4, 10, 27, 72, 187, 488}},
		{seq: Pratt, length: 50000, want: []int{1, 2, 3, 4, 6, 8, 9, 12, 16}},
		{seq: Pratt25, length: 50000, want: []int{1, 2, 4, 5, 8, 10, 16, 20}},
		{seq: Pratt34, length: 50000, want: []int{1, 3, 4, 9, 12, 16, 27}},
		{seq: Sedgewick86, length: 50000, want: []int{1, 5, 19, 41, 109}},
		{seq: Sedgewick82, length: 50000, want: []int{1, 8, 23, 77, 281}},
		{seq: Hibbard, length: 50000, want: []int{1, 3, 7, 15, 31}},
		{seq: PapernovStasevich, length: 50000, want: []int{1, 3, 5, 9, 17, 33}},
		{seq: IncerpiSedgewick, length: 50000, want: []int{1, 3, 7, 21, 48, 112}},

		// Derived from N, so the whole sequence is asserted at one size.
		{
			seq: FrankLazarus, length: 1000, exact: true,
			want: []int{1, 3, 7, 15, 31, 63, 125, 251, 501},
		},
		{
			seq: GonnetBaezaYates, length: 1000, exact: true,
			want: []int{1, 3, 8, 19, 42, 93, 206, 454},
		},
		{
			seq: Shell, length: 1000, exact: true,
			want: []int{1, 3, 7, 15, 31, 62, 125, 250, 500},
		},
	}

	require.Len(t, cases, len(allSequences),
		"every catalog entry needs golden terms")

	for _, tc := range cases {
		t.Run(tc.seq.Name, func(t *testing.T) {
			t.Parallel()
			got := tc.seq.Gaps(tc.length)
			if tc.exact {
				require.Equal(t, tc.want, got)
				return
			}
			require.GreaterOrEqual(t, len(got), len(tc.want),
				"too few gaps at length %d: %v", tc.length, got)
			require.Equal(t, tc.want, got[:len(tc.want)])
		})
	}
}

// TestCiuraExtension covers the one part of the catalog with no source to
// check against: the tables continue past their last searched term by the
// unvalidated 2.25 ratio convention, and that continuation must still satisfy
// the contract rather than stalling or repeating the last term.
func TestCiuraExtension(t *testing.T) {
	t.Parallel()
	cases := []struct {
		seq  GapSequence
		last int // last tabulated term
	}{
		{seq: Ciura, last: 1750},
		{seq: Ciura128, last: 126},
		{seq: Ciura1000, last: 995},
	}
	for _, tc := range cases {
		t.Run(tc.seq.Name, func(t *testing.T) {
			t.Parallel()
			gaps := tc.seq.Gaps(1 << 20)
			i := 0
			for gaps[i] != tc.last {
				i++
				require.Less(t, i, len(gaps), "last tabulated term missing")
			}
			require.Less(t, i+1, len(gaps), "table was never extended")
			for j := i + 1; j < len(gaps); j++ {
				require.Equal(t, gaps[j-1]*9/4, gaps[j],
					"extension must be the floor of 2.25x the previous gap")
			}
		})
	}
}
