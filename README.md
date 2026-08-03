# Shell Sort in Go

Parallel shell sort using Sedgewick gap sequence. Subarrays with fewer than 5 gaps sort concurrently via goroutines; larger counts sort sequentially.

[![CI](https://github.com/OlegElizarov/Shell_Sort_Golang/actions/workflows/ci.yml/badge.svg)](https://github.com/OlegElizarov/Shell_Sort_Golang/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/OlegElizarov/Shell_Sort_Golang)](https://goreportcard.com/report/github.com/OlegElizarov/Shell_Sort_Golang)
[![codecov](https://codecov.io/gh/OlegElizarov/Shell_Sort_Golang/branch/main/graph/badge.svg)](https://codecov.io/gh/OlegElizarov/Shell_Sort_Golang)
![Go Version](https://img.shields.io/github/go-mod/go-version/OlegElizarov/Shell_Sort_Golang)
[![License](https://img.shields.io/github/license/OlegElizarov/Shell_Sort_Golang)](LICENSE)

## Usage

```go
import "github.com/OlegElizarov/Shell_Sort_Golang/lib"

sorted := lib.ShellSort([]int{5, 3, 8, 1, 2})
```

## Commands

```bash
make build       # Build
make test        # Tests with race detector + coverage
make lint        # golangci-lint (v2+)
make demo        # Run demo
make benchmark   # Benchmarks + CPU/mem profiles
make ci          # Full CI check (build + vet + test + lint)
```

## Profiling

```bash
make benchmark
go tool pprof -alloc_objects -http=:8080 mem
```

## Gap Sequences

- **Sedgewick** (active) — `selectStepSedgewick`
- **Hibbard** (alternative) — `selectStepHibbard`, swap in `lib/sort.go`

See [docs/GAP_SEQUENCES.md](docs/GAP_SEQUENCES.md) for analysis of the other sequences — formulae, complexity bounds, and fit with the parallel design.

## Structure

```
lib/sort.go       — ShellSort + gap sequences + parallel subSort
lib/sort_test.go  — table-driven tests + benchmarks
lib/slices.go     — shared fixture slices
main.go           — demo entrypoint
```
