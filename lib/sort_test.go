package lib

import (
	"math/rand/v2"
	"runtime"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

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
	length := rand.IntN(50000)
	d := selectStepSedgewick(length)
	require.True(t, d[0]*3 < length)
	require.Equal(t, 1, d[0])
}

func TestSelectStepHibbard(t *testing.T) {
	t.Parallel()
	length := rand.IntN(100)
	d := selectStepHibbard(length)
	require.Equal(t, 1, d[0])
}

func BenchmarkshellSort(i int, b *testing.B) {
	runtime.GOMAXPROCS(4)
	rng := rand.New(rand.NewPCG(1, 2))
	slice := rng.Perm(i)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ShellSort(slice)
		}
	})
}

func BenchmarkstdSort(i int, b *testing.B) {
	rng := rand.New(rand.NewPCG(1, 2))
	slice := rng.Perm(i)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			slices.Sort(slice)
		}
	})
}

//func BenchmarkShellSort1(b *testing.B) { BenchmarkshellSort(1000, b) }
//func BenchmarkShellSort1(b *testing.B) { BenchmarkshellSort(10000, b) }
//func BenchmarkShellSort2(b *testing.B) { BenchmarkshellSort(20000, b) }
func BenchmarkShellSort3(b *testing.B) { BenchmarkshellSort(30000, b) }

//func BenchmarkShellSort4(b *testing.B) { BenchmarkshellSort(40000, b) }
//func BenchmarkShellSort5(b *testing.B) { BenchmarkshellSort(50000, b) }
func BenchmarkStdSort(b *testing.B) { BenchmarkstdSort(30000, b) }
