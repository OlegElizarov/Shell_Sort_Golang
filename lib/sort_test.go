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

// randomSlice returns a deterministic permutation of 0..n-1. Seeding
// explicitly per call site keeps every case reproducible from its own two
// constants, which the previous rand.IntN-based lengths were not.
func randomSlice(tb testing.TB, n int, seed1, seed2 uint64) []int {
	tb.Helper()
	return rand.New(rand.NewPCG(seed1, seed2)).Perm(n)
}

// ascendingSlice returns 0..n-1 in order.
func ascendingSlice(n int) []int {
	s := make([]int, n)
	for i := range s {
		s[i] = i
	}
	return s
}

// descendingSlice returns n-1..0, the worst case for a plain insertion sort.
func descendingSlice(n int) []int {
	s := ascendingSlice(n)
	slices.Reverse(s)
	return s
}

// sortCase is one input shape. Cases are shared between TestShellSort and
// TestShellSortAllSequences so that adding a shape covers the default sequence
// and all 18 catalog sequences at once.
type sortCase struct {
	Name  string
	Input []int
}

// sortCases covers the boundaries (nil, empty, single) and the three orderings
// whose costs differ: random, already sorted, and reversed. Sizes straddle the
// gap sequences' early terms so that short inputs exercise the one-pass path
// and longer ones exercise several passes.
func sortCases(tb testing.TB) []sortCase {
	tb.Helper()
	return []sortCase{
		{Name: "nil", Input: nil},
		{Name: "empty", Input: []int{}},
		{Name: "single", Input: []int{42}},
		{Name: "random 16", Input: randomSlice(tb, 16, 1, 2)},
		{Name: "random 100", Input: randomSlice(tb, 100, 3, 4)},
		{Name: "random 500", Input: randomSlice(tb, 500, 5, 6)},
		{Name: "already sorted 100", Input: ascendingSlice(100)},
		{Name: "reverse sorted 100", Input: descendingSlice(100)},
	}
}

// requireSorts checks one input against the stdlib oracle. Both sides clone the
// same input, so the oracle is compared against untouched data — the previous
// version of this test asserted against a hardcoded literal and then called
// ShellSort a second time on the slice the first call had already sorted in
// place, which made the second assertion true no matter what ShellSort did.
func requireSorts(t *testing.T, input []int, opts ...Option) {
	t.Helper()

	want := slices.Clone(input)
	slices.Sort(want)

	got := ShellSort(slices.Clone(input), opts...)

	require.Equal(t, want, got)
}

func TestShellSort(t *testing.T) {
	t.Parallel()
	for _, tc := range sortCases(t) {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			requireSorts(t, tc.Input)
		})
	}
}

// TestShellSortAllSequences runs the full case table against every sequence in
// the catalog. This is what the GapSequence contract buys: one loop covers all
// 18 implementations, so a generator that violates the contract in a way the
// contract test misses still surfaces here as a mis-sorted slice.
func TestShellSortAllSequences(t *testing.T) {
	t.Parallel()
	cases := sortCases(t)
	for _, seq := range allSequences {
		t.Run(seq.Name, func(t *testing.T) {
			t.Parallel()
			for _, tc := range cases {
				t.Run(tc.Name, func(t *testing.T) {
					t.Parallel()
					requireSorts(t, tc.Input, WithGapSequence(seq))
				})
			}
		})
	}
}

// TestShellSortSortsInPlace pins the documented aliasing contract: the
// returned slice is the argument, already sorted, not a copy.
func TestShellSortSortsInPlace(t *testing.T) {
	t.Parallel()
	in := randomSlice(t, 64, 7, 8)

	got := ShellSort(in)

	require.True(t, slices.IsSorted(in), "argument must be sorted in place")
	require.Equal(t, in, got)
}

// FuzzShellSort checks ShellSort against the stdlib oracle on inputs with
// duplicates. Values are drawn from a small range on purpose: Perm produces a
// permutation, so every table case above has distinct elements and cannot
// exercise ties at all.
func FuzzShellSort(f *testing.F) {
	f.Add(int64(1), uint8(0))
	f.Add(int64(2), uint8(1))
	f.Add(int64(3), uint8(37))
	f.Add(int64(4), uint8(255))

	f.Fuzz(func(t *testing.T, seed int64, n uint8) {
		rng := rand.New(rand.NewPCG(uint64(seed), uint64(n)))
		in := make([]int, n)
		for i := range in {
			in[i] = rng.IntN(8)
		}

		requireSorts(t, in)
	})
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

// TestShellSortWithGapSequence covers the parts of the option API that the
// catalog loop in TestShellSortAllSequences does not: a sequence supplied by
// the caller rather than taken from the catalog, and option precedence.
func TestShellSortWithGapSequence(t *testing.T) {
	t.Parallel()
	input := randomSlice(t, 500, 11, 13)

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
		requireSorts(t, input, WithGapSequence(mine))
	})

	t.Run("last option wins", func(t *testing.T) {
		t.Parallel()
		requireSorts(t, input, WithGapSequence(Shell), WithGapSequence(Ciura))
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
