package main

import (
	"Shell_Sort_Golang/lib"
	"fmt"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("Hello")
	fmt.Println(lib.ShellSort(lib.IntSliceShort))
	fmt.Println(lib.ShellSort(lib.IntSliceBig))
}
