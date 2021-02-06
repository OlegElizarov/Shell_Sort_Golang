# Shell_Sort_Golang
Simple implementation of Shell sort in Golang

![Tests](https://github.com/OlegElizarov/Shell_Sort_Golang/workflows/Tests/badge.svg)

GET PROFILES: go test -bench=. -benchmem  -memprofile=mem -cpuprofile=cpu . <br>
PPROF SERVER FOR MEM PROFILE: go tool pprof -alloc_objects -http=:8080 mem <br>
PPROF TERMINAL FOR MEM PROFILE: go tool pprof -alloc_objects mem <br>