package lib

import (
	"github.com/stretchr/testify/require"
	"math/rand"
	"sort"
	"testing"
	"time"
)

func TestShellSort(t *testing.T) {
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
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, ShellSort(tc.Input), tc.Expected)
			require.True(t, sort.IsSorted(sort.IntSlice(ShellSort(tc.Input))))
		})
	}
}

// func TestShellSort1(t *testing.T) {
//	require.Equal(t, ShellSort(IntSliceShort), IntSliceCorrect)
//	require.True(t, sort.IsSorted(sort.IntSlice(ShellSort(IntSliceShort))))
//}
//
// func TestShellSort2(t *testing.T) {
//	require.True(t, sort.IsSorted(sort.IntSlice(ShellSort(IntSliceBig))))
//}
//
// func TestShellSort3(t *testing.T) {
//	fmt.Println("500 numbers arr")
//	require.Equal(t, ShellSort(IntSliceVeryBig), IntSliceVeryBigCorrect)
//}

func TestSelectStepSedgewick(t *testing.T) {
	length := rand.Intn(100)
	d := selectStepSedgewick(length)
	require.True(t, d[0]*3 < length)
	require.Equal(t, d[(len(d))-1], 1)
}

func TestSelectStepHibbard(t *testing.T) {
	length := rand.Intn(100)
	d := selectStepHibbard(length)
	//require.True(t, d[0]*3 < length)
	require.Equal(t, d[(len(d))-1], 1)
}

func BenchmarkshellSort(i int, b *testing.B) {
	rand.Seed(time.Now().Unix())
	slice := rand.Perm(i)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ShellSort(slice)
		}
	})
}

func BenchmarkShellSort1(b *testing.B) { BenchmarkshellSort(10000, b) }
func BenchmarkShellSort2(b *testing.B) { BenchmarkshellSort(20000, b) }
func BenchmarkShellSort3(b *testing.B) { BenchmarkshellSort(30000, b) }
func BenchmarkShellSort4(b *testing.B) { BenchmarkshellSort(40000, b) }
func BenchmarkShellSort5(b *testing.B) { BenchmarkshellSort(50000, b) }
