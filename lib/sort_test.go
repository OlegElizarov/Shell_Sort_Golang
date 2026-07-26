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

func TestSelectStepSedgewick(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		length int
	}{
		{name: "zero", length: 0},
		{name: "one", length: 1},
		{name: "small", length: 10},
		{name: "medium", length: 500},
		{name: "random", length: rand.IntN(50000) + 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d := selectStepSedgewick(tc.length)
			require.NotEmpty(t, d)
			require.Equal(t, 1, d[0])
			if tc.length > 1 {
				require.True(t, d[0]*3 < tc.length)
			}
		})
	}
}

func TestSelectStepHibbard(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		length int
	}{
		{name: "zero", length: 0},
		{name: "one", length: 1},
		{name: "small", length: 10},
		{name: "medium", length: 500},
		{name: "random", length: rand.IntN(100) + 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d := selectStepHibbard(tc.length)
			require.NotEmpty(t, d)
			require.Equal(t, 1, d[0])
		})
	}
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
