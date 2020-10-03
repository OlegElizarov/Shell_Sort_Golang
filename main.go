package main

import (
	"Shell_Sort_Golang/lib"
	"fmt"
)

func main() {
	fmt.Println("Hello")
	fmt.Println(lib.ShellSort(lib.IntSliceShort))
	fmt.Println(lib.ShellSort(lib.IntSliceBig))

}
