.PHONY: build vet test test-short lint lint-fix demo benchmark bench-smoke \
	bench-record bench-baseline bench-stat benchstat-install profile-view clean help ci

# Benchmark knobs. Phase 0 measured +/-13-24% run-to-run spread on ShellSort,
# so 10 counts is not enough to call a winner - 20 is the floor for anything
# compared with benchstat.
BENCH_COUNT ?= 20
BENCH_PKG   ?= ./lib/...
BENCH_BASE  ?= docs/bench-baseline.txt
BENCH_NEW   ?= bench-new.txt

help:
	@echo "Shell Sort Golang - Available targets:"
	@echo "  build            - Build project"
	@echo "  vet              - Run go vet"
	@echo "  test             - Run all tests with coverage"
	@echo "  test-short       - Run specific test TestShellSort"
	@echo "  lint             - Run golangci-lint"
	@echo "  lint-fix         - Auto-fix struct field alignment"
	@echo "  demo             - Run demo"
	@echo "  benchmark        - Run benchmarks with profiling"
	@echo "  bench-smoke      - Run each benchmark once (correctness only, no timings)"
	@echo "  bench-record     - Run benchmarks -count=\$$(BENCH_COUNT) into \$$(BENCH_NEW)"
	@echo "  bench-baseline   - Same, but overwrite \$$(BENCH_BASE) (rewrites the baseline)"
	@echo "  bench-stat       - benchstat \$$(BENCH_BASE) vs \$$(BENCH_NEW)"
	@echo "  benchstat-install- go install golang.org/x/perf/cmd/benchstat@latest"
	@echo "  clean            - Remove build artifacts and profiles"
	@echo "  ci               - Run build, vet, test, lint (matches CI)"

build:
	go build -v ./...

vet:
	go vet ./...

test:
	go test -v -race -coverprofile=coverage.out ./...

test-short:
	go test -run TestShellSort ./lib/...

lint:
	golangci-lint run -c .golangci.yml --timeout=5m

lint-fix:
	go run golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest -fix ./...

demo:
	go run ./cmd

benchmark:
	go test -bench=. -benchmem -memprofile=mem -cpuprofile=cpu ./lib/...
	@echo "\nProfile available. View with:"
	@echo "  go tool pprof -alloc_objects -http=:8080 mem"

# One iteration per benchmark, timings ignored. Cheap enough for CI: proves the
# benchmark code compiles, runs and stays race-free without pretending a shared
# runner can produce a meaningful number.
bench-smoke:
	go test -run '^$$' -bench=. -benchtime=1x $(BENCH_PKG)

benchstat-install:
	go install golang.org/x/perf/cmd/benchstat@latest

bench-record:
	@command -v benchstat >/dev/null || { echo "benchstat missing - run: make benchstat-install"; exit 1; }
	go test -bench=. -benchmem -count=$(BENCH_COUNT) $(BENCH_PKG) | tee $(BENCH_NEW)
	@echo "\nWrote $(BENCH_NEW). Compare with: make bench-stat"

bench-baseline:
	@command -v benchstat >/dev/null || { echo "benchstat missing - run: make benchstat-install"; exit 1; }
	go test -bench=. -benchmem -count=$(BENCH_COUNT) $(BENCH_PKG) | tee $(BENCH_BASE)
	@echo "\nRewrote baseline $(BENCH_BASE)."

bench-stat:
	@command -v benchstat >/dev/null || { echo "benchstat missing - run: make benchstat-install"; exit 1; }
	@test -f $(BENCH_BASE) || { echo "$(BENCH_BASE) missing - run: make bench-baseline"; exit 1; }
	@test -f $(BENCH_NEW) || { echo "$(BENCH_NEW) missing - run: make bench-record"; exit 1; }
	benchstat $(BENCH_BASE) $(BENCH_NEW)

profile-view:
	go tool pprof -alloc_objects -http=:8080 mem

clean:
	go clean
	rm -f coverage.out mem cpu $(BENCH_NEW)

ci: build vet test lint
	@echo "CI checks passed"
