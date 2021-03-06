package main

import (
	"fmt"
	"github.com/OlegElizarov/Shell_Sort_Golang/lib"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("Hello")
	fmt.Println(lib.ShellSort(lib.IntSliceShort))
	fmt.Println(lib.ShellSort(lib.IntSliceBig))
}
