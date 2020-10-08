package lib

import (
	"math"
	"sort"
	"sync"
)

func ShellSort(slice []int) []int {
	//runtime.GOMAXPROCS(2)
	length := len(slice)
	var wg sync.WaitGroup
	d := selectStepSedgewick(length)
	//d := selectStepHibbard(length)
	for _, step := range d {
		//step == num of subarrays
		if step < 5 { // 5 sub array can't sort async good enough because of cores num
			wg.Add(step)
			for i := 0; i < step; i++ {
				go subSort(&slice, step, i, &wg)
			}
			wg.Wait()
		} else {
			for i := 0; i < step; i++ {
				subSort(&slice, step, i, &wg)
			}
		}
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
	if step < 5 {
		defer wg.Done()
	}

	// Sort elements inside subslice

	//bubble
	//for j := position; j+step < len(*slice); j += step {
	//	for z := position; z+step < len(*slice)-j; z += step {
	//		if (*slice)[z] > (*slice)[z+step] {
	//			(*slice)[z+step], (*slice)[z] = (*slice)[z], (*slice)[z+step]
	//		}
	//	}
	//}

	//insertionSort
	var temp int
	var item int // previous elem index
	for counter := position; counter < len((*slice)); counter += step {
		temp = (*slice)[counter]
		item = counter - step
		for item >= 0 && (*slice)[item] > temp {
			(*slice)[item+step] = (*slice)[item]
			(*slice)[item] = temp
			item -= step
		}
	}
}
