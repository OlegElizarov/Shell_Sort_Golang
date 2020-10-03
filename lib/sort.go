package lib

import (
	"math"
	"sort"
	"sync"
)

func ShellSort(slice []int) []int {
	length := len(slice)
	var wg sync.WaitGroup
	d := selectStepSedgewick(length)
	for _, step := range d {
		wg.Add(step)
		for i := 0; i < step; i++ {
			go subSort(&slice, step, i, &wg)
		}
		wg.Wait()
	}
	return slice
}

func selectStepHibbard(len int) (steps []int) {
	var d float64
	for i := 1; d < float64(len); i++ {
		d = math.Pow(2, float64(i)) - 1
		if d < float64(len) {
			steps = append(steps, int(d))
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(steps)))
	return
}

func selectStepSedgewick(len int) (steps []int) {
	var d float64
	for i := 0.0; d*3 < float64(len); i++ {
		if int(i)%2 == 0 {
			d = 9*(math.Pow(2.0, i)) - 9*(math.Pow(2.0, i/2)) + 1
		} else {
			d = 8*(math.Pow(2.0, i)) - 6*(math.Pow(2.0, (i+1)/2)) + 1
		}
		if d*3 < float64(len) {
			steps = append(steps, int(d))
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(steps)))
	return
}

func subSort(slice *[]int, step, position int, wg *sync.WaitGroup) {
	defer wg.Done()
	// Sort elements inside subslice
	for j := position; j+step < len(*slice); j += step {
		for z := position; z+step < len(*slice)-j; z += step {
			if (*slice)[z] > (*slice)[z+step] {
				(*slice)[z+step], (*slice)[z] = (*slice)[z], (*slice)[z+step]
			}
		}
	}
}
