package lib

import (
	"math/rand/v2"
	"runtime"
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

// benchGOMAXPROCS pins both benchmarks to the same parallelism so the
// comparison isn't at the mercy of whatever GOMAXPROCS the environment
// happens to default to on a given machine or CI runner.
var benchGOMAXPROCS = runtime.NumCPU()

func TestShellSort(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name     string
		Input    []int
		Expected []int
	}{
		{Name: "Short slice test", Input: IntSliceShort, Expected: IntSliceCorrect},
		{Name: "Big slice test", Input: IntSliceBig, Expected: IntSliceBigCorrect},
		{Name: "Very big slice test", Input: IntSliceVeryBig, Expected: IntSliceVeryBigCorrect},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, ShellSort(tc.Input), tc.Expected)
			require.True(t, slices.IsSorted(ShellSort(tc.Input)))
		})
	}
}

// TestShellSortGenericElements instantiates ShellSort at element types other
// than int. The point is the compile: type inference has to keep working for
// the existing []int call sites while also accepting anything cmp.Ordered.
func TestShellSortGenericElements(t *testing.T) {
	t.Parallel()

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		got := ShellSort([]string{"pear", "apple", "fig", "date"})
		require.Equal(t, []string{"apple", "date", "fig", "pear"}, got)
	})

	t.Run("float64", func(t *testing.T) {
		t.Parallel()
		got := ShellSort([]float64{2.5, -1, 0, 100.25, -1.5})
		require.Equal(t, []float64{-1.5, -1, 0, 2.5, 100.25}, got)
	})

	t.Run("named slice type", func(t *testing.T) {
		t.Parallel()
		// ~[]E means a defined type with a slice underlying type still
		// satisfies the constraint, and comes back as its own type. The
		// return type is asserted at compile time by this signature: it
		// would not build if ShellSort returned []int here.
		type ranking []int
		sorted := func(r ranking) ranking { return ShellSort(r) }
		require.Equal(t, ranking{1, 2, 3}, sorted(ranking{3, 1, 2}))
	})
}

// TestShellSortWithGapSequence checks that the option actually selects the
// sequence, across the whole catalog and a caller-supplied one. Phase 5
// replaces the fixture-based table above; this covers the new API in the
// meantime.
func TestShellSortWithGapSequence(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewPCG(11, 13))
	input := rng.Perm(500)
	want := slices.Clone(input)
	slices.Sort(want)

	for _, seq := range allSequences {
		t.Run(seq.Name, func(t *testing.T) {
			t.Parallel()
			got := ShellSort(slices.Clone(input), WithGapSequence(seq))
			require.Equal(t, want, got)
		})
	}

	t.Run("caller-supplied", func(t *testing.T) {
		t.Parallel()
		// Any sequence meeting the GapSequence contract sorts correctly;
		// only performance depends on the choice.
		mine := GapSequence{
			Name: "mine",
			Gaps: func(length int) []int {
				gaps := []int{1}
				for h := 7; h < length; h *= 3 {
					gaps = append(gaps, h)
				}
				return gaps
			},
		}
		got := ShellSort(slices.Clone(input), WithGapSequence(mine))
		require.Equal(t, want, got)
	})

	t.Run("last option wins", func(t *testing.T) {
		t.Parallel()
		got := ShellSort(slices.Clone(input),
			WithGapSequence(Shell), WithGapSequence(Ciura))
		require.Equal(t, want, got)
	})
}

// benchSizes cover small, medium, and large inputs so ShellSort vs slices.Sort
// scale comparisons are visible across the whole range, not just one point.
var benchSizes = []int{1000, 10000, 30000}

// benchInputs caches one base permutation per size so ShellSort and
// slices.Sort benchmarks clone from the exact same underlying slice instead
// of each generating their own copy from a matching seed — same object, not
// just same seed replayed twice.
var benchInputs = map[int][]int{}

func benchInput(size int) []int {
	if base, ok := benchInputs[size]; ok {
		return base
	}
	rng := rand.New(rand.NewPCG(1, 2))
	base := rng.Perm(size)
	benchInputs[size] = base
	return base
}

// benchmarkSort times fn against a fresh clone of the shared base permutation
// on every iteration. Reusing one slice across b.N iterations (the old
// approach) would sort it once and leave every subsequent iteration timing an
// already-sorted best case, which silently favors whichever algorithm runs
// first and makes the two benchmarks incomparable.
func benchmarkSort(b *testing.B, size int, fn func([]int)) {
	b.Helper()
	runtime.GOMAXPROCS(benchGOMAXPROCS)
	base := benchInput(size)
	b.ReportAllocs()
	for b.Loop() {
		b.StopTimer()
		slice := slices.Clone(base)
		b.StartTimer()
		fn(slice)
	}
}

func BenchmarkShellSort(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			benchmarkSort(b, size, func(s []int) { ShellSort(s) })
		})
	}
}

func BenchmarkStdSort(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			benchmarkSort(b, size, slices.Sort)
		})
	}
}
