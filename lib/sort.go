// Package lib provides a Shell sort with a pluggable gap sequence and
// goroutine-parallelized passes.
//
// The gap sequence is a runtime choice rather than a compile-time one. The
// catalog in gaps.go covers the published sequences; callers can also supply
// their own generator, since GapSequence is an ordinary struct:
//
//	sorted := lib.ShellSort(x)                                  // Tokuda, the default
//	sorted := lib.ShellSort(x, lib.WithGapSequence(lib.Ciura))  // a catalog entry
//	sorted := lib.ShellSort(x, lib.WithGapSequence(mine))       // your own
package lib

import (
	"cmp"
	"slices"
	"sync"
)

// DefaultGapSequence is the sequence ShellSort uses when none is given.
//
// Tokuda is the default because it is closed-form and unbounded (no table to
// run out of), needs ~17 passes at N = 10^6, ties every Ciura variant on
// comparisons at N = 10^4 while using the fewest exchanges of the practical
// sequences, and scales its first gap with the input size — which matters more
// here than in a sequential implementation, since the first gap sets how much
// parallel width the widest pass has. See docs/GAP_SEQUENCES.md §7.
var DefaultGapSequence = Tokuda

// parallelSubarrayCeiling is the pass width below which the subarrays are
// sorted concurrently.
//
// This is the original implementation's rule, preserved here unchanged so that
// this phase of the rewrite is behaviour-preserving apart from the change of
// default sequence. It is also backwards: a pass with fewer subarrays has less
// parallelism available and longer subarrays, so this parallelizes the cheap
// tail passes and runs the wide early passes sequentially. Replacing it with a
// threshold on work per subarray is a separate change, with its own before and
// after measurements — see phase 7 of docs/MODERNIZATION_PLAN.md.
const parallelSubarrayCeiling = 5

// config holds the resolved options for one ShellSort call.
type config struct {
	sequence GapSequence
}

// Option configures a ShellSort call. Options are applied in order, so a later
// one overrides an earlier one.
type Option func(*config)

// WithGapSequence sets the gap sequence. Any entry from the catalog in gaps.go
// works, as does a caller-supplied GapSequence whose Gaps function meets the
// contract documented on that type.
func WithGapSequence(sequence GapSequence) Option {
	return func(cfg *config) {
		cfg.sequence = sequence
	}
}

// ShellSort sorts x in ascending order and returns it. The sort is in place:
// the returned slice shares its backing array with the argument.
//
// Passes run from the largest gap down to 1. Narrow passes have their
// subarrays sorted concurrently; see parallelSubarrayCeiling.
func ShellSort[S ~[]E, E cmp.Ordered](x S, opts ...Option) S {
	cfg := config{sequence: DefaultGapSequence}
	for _, opt := range opts {
		opt(&cfg)
	}

	var wg sync.WaitGroup
	for _, gap := range slices.Backward(cfg.sequence.Gaps(len(x))) {
		// gap is both the stride and the number of subarrays in this pass.
		if gap < parallelSubarrayCeiling {
			wg.Add(gap)
			for start := range gap {
				go func() {
					defer wg.Done()
					subSort(x, gap, start)
				}()
			}
			wg.Wait()
			continue
		}
		for start := range gap {
			subSort(x, gap, start)
		}
	}
	return x
}

// subSort insertion-sorts the subarray of x that begins at start and steps by
// gap, leaving the rest of x untouched. Concurrent calls with the same gap and
// different starts touch disjoint elements, so they are safe to run together.
func subSort[E cmp.Ordered](x []E, gap, start int) {
	for pos := start; pos < len(x); pos += gap {
		temp := x[pos]
		item := pos - gap
		for item >= 0 && x[item] > temp {
			x[item+gap], x[item] = x[item], temp
			item -= gap
		}
	}
}
