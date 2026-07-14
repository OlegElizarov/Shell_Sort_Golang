.PHONY: build vet test test-short lint lint-fix demo benchmark clean help

help:
	@echo "Shell Sort Golang - Available targets:"
	@echo "  build       - Build project"
	@echo "  vet         - Run go vet"
	@echo "  test        - Run all tests with coverage"
	@echo "  test-short  - Run specific test TestShellSort"
	@echo "  lint        - Run golangci-lint"
	@echo "  lint-fix    - Auto-fix struct field alignment"
	@echo "  demo        - Run demo"
	@echo "  benchmark   - Run benchmarks with profiling"
	@echo "  clean       - Remove build artifacts and profiles"
	@echo "  ci          - Run build, vet, test (matches CI)"

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
	go run .

benchmark:
	go test -bench=. -benchmem -memprofile=mem -cpuprofile=cpu ./lib/...
	@echo "\nProfile available. View with:"
	@echo "  go tool pprof -alloc_objects -http=:8080 mem"

profile-view:
	go tool pprof -alloc_objects -http=:8080 mem

clean:
	go clean
	rm -f coverage.out mem cpu

ci: build vet test lint
	@echo "CI checks passed"
