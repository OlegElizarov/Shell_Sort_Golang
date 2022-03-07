package lib

import (
	"math"
	"sync"
)

//TODO:const for choosing select func(Sedgewick/Hibbard)

// ShellSort ...
func ShellSort(slice []int) []int {
	length := len(slice)
	var wg sync.WaitGroup
	d := selectStepSedgewick(length)
	//d := selectStepHibbard(length)
	for ind := range d {
		//step == num of subarrays
		if d[len(d)-ind-1] < 5 { // 5 sub array can't sort async good enough because of cores num
			wg.Add(d[len(d)-ind-1])
			for i := 0; i < d[len(d)-ind-1]; i++ {
				go subSort(&slice, d[len(d)-ind-1], i, &wg)
			}
			wg.Wait()
		} else {
			for i := 0; i < d[len(d)-ind-1]; i++ {
				subSort(&slice, d[len(d)-ind-1], i, &wg)
			}
		}
	}
	return slice
}

func selectStepSedgewick(length int) []int {
	var d float64
	steps := make([]int, 10)

	for i := 0.0; d*3 < float64(length); i++ {
		if int(i)%2 == 0 {
			d = 9*(math.Pow(2.0, i)) - 9*(math.Pow(2.0, i/2)) + 1
		} else {
			d = 8*(math.Pow(2.0, i)) - 6*(math.Pow(2.0, (i+1)/2)) + 1
		}
		if d*3 < float64(length) {
			if int(i) < len(steps) {
				steps[int(i)] = int(d)
			} else {
				steps = append(steps, int(d))
			}
		}
	}
	//steps = steps[:int(i)-1]
	return steps
}

func selectStepHibbard(length int) []int {
	var d float64
	steps := make([]int, 10)
	for i := 1; d < float64(length); i++ {
		d = math.Pow(2, float64(i)) - 1
		if d < float64(length) {
			if i < len(steps) {
				steps[i-1] = int(d)
			} else {
				steps = append(steps, int(d))
			}
		}
	}
	//steps = steps[len(steps)-i+2:]
	return steps
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
	for counter := position; counter < len(*slice); counter += step {
		temp = (*slice)[counter]
		item = counter - step
		for item >= 0 && (*slice)[item] > temp {
			//(*slice)[item+step] = (*slice)[item]
			//(*slice)[item] = temp
			(*slice)[item+step], (*slice)[item] = (*slice)[item], temp
			item -= step
		}
	}
}
