package main

import (
	"fmt"
	"runtime"

	"github.com/OlegElizarov/Shell_Sort_Golang/lib"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println(lib.ShellSort(lib.IntSliceShort))
	fmt.Println(lib.ShellSort(lib.IntSliceBig))
}
