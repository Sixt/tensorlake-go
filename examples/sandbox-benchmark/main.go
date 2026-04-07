// Copyright 2026 SIXT SE
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Command sandbox-benchmark stress-tests sandbox creation latency by
// running a series of concurrency levels, launching N sandboxes at each
// level and measuring create API response time and time-to-running.
//
// Each concurrency level is repeated R times (default 1). The summary
// reports mean±std across repeats for each percentile.
//
// Usage:
//
//	export TENSORLAKE_API_KEY=<your-api-key>
//	go run ./examples/sandbox-benchmark -con 1,10,100
//	go run ./examples/sandbox-benchmark -repeat 5 -con 1,10,100,1000
//
// Flags:
//
//	-con       Comma-separated concurrency levels (default: 1,10,100)
//	-repeat    Number of times to repeat each concurrency level (default: 1)
//	-timeout   Sandbox timeout in seconds (default: 120)
//	-poll      Poll interval for status checks (default: 100ms)
//	-max-wait  Maximum wait time per sandbox before giving up (default: 120s)
package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	tensorlake "github.com/sixt/tensorlake-go"
)

type result struct {
	index      int
	sandboxID  string
	createDur  time.Duration // API call: create request → response (sandbox is pending)
	pendingDur time.Duration // pending → running (excludes create API time)
	readyDur   time.Duration // total: create request → running (createDur + pendingDur)
	err        error
}

// runStats holds the percentile values from a single run at a given concurrency level.
type runStats struct {
	concurrency int
	succeeded   int
	failed      int
	totalDur    time.Duration

	// Percentile values extracted from this run's sorted durations.
	createPcts  map[string]time.Duration
	pendingPcts map[string]time.Duration
	readyPcts   map[string]time.Duration
}

var pctDefs = []struct {
	name string
	p    float64
}{
	{"Min", 0},
	{"P50", 0.50}, {"P75", 0.75}, {"P90", 0.90},
	{"P95", 0.95}, {"P99", 0.99}, {"P99.9", 0.999},
	{"Max", 1.0},
}

func main() {
	con := flag.String("con", "1,10,100", "comma-separated concurrency levels")
	repeat := flag.Int("repeat", 1, "number of times to repeat each concurrency level")
	timeout := flag.Int64("timeout", 120, "sandbox timeout in seconds")
	pollInterval := flag.Duration("poll", 100*time.Millisecond, "poll interval for status checks")
	maxWait := flag.Duration("max-wait", 120*time.Second, "maximum wait time per sandbox")
	flag.Parse()

	var levels []int
	for _, part := range strings.Split(*con, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil || n <= 0 {
			fmt.Fprintf(os.Stderr, "invalid concurrency level: %q\n", part)
			os.Exit(1)
		}
		levels = append(levels, n)
	}

	apiKey := os.Getenv("TENSORLAKE_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "TENSORLAKE_API_KEY environment variable is required")
		os.Exit(1)
	}

	c := tensorlake.NewClient(tensorlake.WithAPIKey(apiKey))
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// allRuns[levelIdx] = slice of runStats, one per repeat.
	allRuns := make([][]runStats, len(levels))

	runIdx := 0
	totalRuns := len(levels) * *repeat
	for li, n := range levels {
		for r := range *repeat {
			if ctx.Err() != nil {
				break
			}
			runIdx++
			fmt.Fprintf(os.Stderr, "\n=== Run %d/%d: concurrency=%d repeat=%d/%d ===\n",
				runIdx, totalRuns, n, r+1, *repeat)
			stats := runLevel(ctx, c, n, *timeout, *pollInterval, *maxWait)
			allRuns[li] = append(allRuns[li], stats)
		}
	}

	fmt.Printf("\n")
	if *repeat == 1 {
		// Single repeat: show raw percentile values.
		var singleRuns []runStats
		for _, runs := range allRuns {
			if len(runs) > 0 {
				singleRuns = append(singleRuns, runs[0])
			}
		}
		printSummarySingle(singleRuns)
	} else {
		// Multiple repeats: show mean±std across repeats.
		printSummaryRepeated(allRuns)
	}
}

func runLevel(ctx context.Context, c *tensorlake.Client, n int, timeout int64, pollInterval, maxWait time.Duration) runStats {
	fmt.Fprintf(os.Stderr, "Launching %d sandboxes...\n", n)
	start := time.Now()

	results := make([]result, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = bench(ctx, c, idx, timeout, pollInterval, maxWait)
		}(i)
	}
	wg.Wait()
	totalDur := time.Since(start)

	// Clean up.
	fmt.Fprintf(os.Stderr, "Cleaning up %d sandboxes...\n", n)
	var cleanWg sync.WaitGroup
	for _, r := range results {
		if r.sandboxID == "" {
			continue
		}
		cleanWg.Add(1)
		go func(id string) {
			defer cleanWg.Done()
			cleanCtx, cleanCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cleanCancel()
			_ = c.DeleteSandbox(cleanCtx, id)
		}(r.sandboxID)
	}
	cleanWg.Wait()

	// Collect and sort durations.
	stats := runStats{
		concurrency: n,
		totalDur:    totalDur,
		createPcts:  make(map[string]time.Duration),
		pendingPcts: make(map[string]time.Duration),
		readyPcts:   make(map[string]time.Duration),
	}
	var createDurs, pendingDurs, readyDurs []time.Duration
	for _, r := range results {
		if r.err != nil {
			stats.failed++
			fmt.Fprintf(os.Stderr, "  [%3d] FAIL: %v\n", r.index, r.err)
		} else {
			stats.succeeded++
			createDurs = append(createDurs, r.createDur)
			pendingDurs = append(pendingDurs, r.pendingDur)
			readyDurs = append(readyDurs, r.readyDur)
		}
	}

	slices.SortFunc(createDurs, func(a, b time.Duration) int { return int(a - b) })
	slices.SortFunc(pendingDurs, func(a, b time.Duration) int { return int(a - b) })
	slices.SortFunc(readyDurs, func(a, b time.Duration) int { return int(a - b) })

	// Extract percentiles.
	for _, p := range pctDefs {
		stats.createPcts[p.name] = pctValue(createDurs, p.p, p.name)
		stats.pendingPcts[p.name] = pctValue(pendingDurs, p.p, p.name)
		stats.readyPcts[p.name] = pctValue(readyDurs, p.p, p.name)
	}
	// Also store mean and stddev.
	stats.createPcts["Mean"] = meanDur(createDurs)
	stats.createPcts["Stddev"] = stddevDur(createDurs)
	stats.pendingPcts["Mean"] = meanDur(pendingDurs)
	stats.pendingPcts["Stddev"] = stddevDur(pendingDurs)
	stats.readyPcts["Mean"] = meanDur(readyDurs)
	stats.readyPcts["Stddev"] = stddevDur(readyDurs)

	fmt.Fprintf(os.Stderr, "Done: %d succeeded, %d failed in %s\n",
		stats.succeeded, stats.failed, totalDur.Round(time.Millisecond))

	return stats
}

func bench(ctx context.Context, c *tensorlake.Client, idx int, timeout int64, pollInterval, maxWait time.Duration) result {
	r := result{index: idx}

	start := time.Now()
	resp, err := c.CreateSandbox(ctx, &tensorlake.CreateSandboxRequest{
		TimeoutSecs: &timeout,
	})
	r.createDur = time.Since(start)

	if err != nil {
		r.err = fmt.Errorf("create: %w", err)
		return r
	}
	r.sandboxID = resp.SandboxId

	if resp.Status == tensorlake.SandboxStatusRunning {
		r.pendingDur = 0
		r.readyDur = r.createDur
		return r
	}

	pendingStart := time.Now()
	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			r.err = ctx.Err()
			return r
		case <-time.After(pollInterval):
		}

		info, err := c.GetSandbox(ctx, r.sandboxID)
		if err != nil {
			r.err = fmt.Errorf("get status: %w", err)
			return r
		}
		if info.Status == tensorlake.SandboxStatusRunning {
			r.pendingDur = time.Since(pendingStart)
			r.readyDur = time.Since(start)
			return r
		}
		if info.Status == tensorlake.SandboxStatusTerminated {
			r.err = fmt.Errorf("sandbox terminated unexpectedly")
			return r
		}
	}

	r.err = fmt.Errorf("timed out waiting for running (%s)", time.Since(start).Round(time.Millisecond))
	return r
}

// printSummarySingle prints raw percentile values when repeat=1.
func printSummarySingle(allRuns []runStats) {
	printHeader(allRuns)

	rows := []string{"Min", "Mean", "Stddev", "P50", "P75", "P90", "P95", "P99", "P99.9", "Max"}

	fmt.Printf("\n--- Create API Response Time (request → sandbox_id) ---\n")
	for _, row := range rows {
		fmt.Printf("%-12s", row)
		for _, s := range allRuns {
			fmt.Printf("  %10s", s.createPcts[row].Round(time.Millisecond))
		}
		fmt.Println()
	}

	fmt.Printf("\n--- Pending to Running (pending → running) ---\n")
	for _, row := range rows {
		fmt.Printf("%-12s", row)
		for _, s := range allRuns {
			fmt.Printf("  %10s", s.pendingPcts[row].Round(time.Millisecond))
		}
		fmt.Println()
	}

	fmt.Printf("\n--- Total Time to Running (request → running) ---\n")
	for _, row := range rows {
		fmt.Printf("%-12s", row)
		for _, s := range allRuns {
			fmt.Printf("  %10s", s.readyPcts[row].Round(time.Millisecond))
		}
		fmt.Println()
	}
}

// printSummaryRepeated prints mean±std across repeats for each percentile.
func printSummaryRepeated(allRuns [][]runStats) {
	// Build a flat header from the first run of each level.
	var header []runStats
	for _, runs := range allRuns {
		if len(runs) > 0 {
			header = append(header, runs[0])
		}
	}
	printHeaderRepeated(header, len(allRuns[0]))

	rows := []string{"Min", "Mean", "Stddev", "P50", "P75", "P90", "P95", "P99", "P99.9", "Max"}

	fmt.Printf("\n--- Create API Response Time (request → sandbox_id) ---\n")
	printRepeatedMetric(allRuns, rows, func(s runStats) map[string]time.Duration { return s.createPcts })

	fmt.Printf("\n--- Pending to Running (pending → running) ---\n")
	printRepeatedMetric(allRuns, rows, func(s runStats) map[string]time.Duration { return s.pendingPcts })

	fmt.Printf("\n--- Total Time to Running (request → running) ---\n")
	printRepeatedMetric(allRuns, rows, func(s runStats) map[string]time.Duration { return s.readyPcts })
}

func printRepeatedMetric(allRuns [][]runStats, rows []string, getPcts func(runStats) map[string]time.Duration) {
	for _, row := range rows {
		fmt.Printf("%-12s", row)
		for _, runs := range allRuns {
			var vals []float64
			for _, s := range runs {
				vals = append(vals, float64(getPcts(s)[row]))
			}
			m := meanF(vals)
			sd := stddevF(vals)
			fmt.Printf("  %8s±%-5s",
				time.Duration(m).Round(time.Millisecond),
				time.Duration(sd).Round(time.Millisecond))
		}
		fmt.Println()
	}
}

func printHeader(allRuns []runStats) {
	fmt.Printf("%-12s", "Concurrency")
	for _, s := range allRuns {
		fmt.Printf("  %10d", s.concurrency)
	}
	fmt.Println()
	fmt.Printf("%-12s", "Succeeded")
	for _, s := range allRuns {
		fmt.Printf("  %10d", s.succeeded)
	}
	fmt.Println()
	fmt.Printf("%-12s", "Failed")
	for _, s := range allRuns {
		fmt.Printf("  %10d", s.failed)
	}
	fmt.Println()
	fmt.Printf("%-12s", "Wall time")
	for _, s := range allRuns {
		fmt.Printf("  %10s", s.totalDur.Round(time.Millisecond))
	}
	fmt.Println()
}

func printHeaderRepeated(header []runStats, repeats int) {
	fmt.Printf("Repeats: %d\n\n", repeats)
	fmt.Printf("%-12s", "Concurrency")
	for _, s := range header {
		fmt.Printf("  %14d", s.concurrency)
	}
	fmt.Println()
}

// pctValue extracts a percentile or min/max from sorted durations.
func pctValue(sorted []time.Duration, p float64, name string) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	switch name {
	case "Min":
		return sorted[0]
	case "Max":
		return sorted[len(sorted)-1]
	default:
		return percentile(sorted, p)
	}
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(p*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func meanDur(durs []time.Duration) time.Duration {
	if len(durs) == 0 {
		return 0
	}
	var sum time.Duration
	for _, d := range durs {
		sum += d
	}
	return sum / time.Duration(len(durs))
}

func stddevDur(durs []time.Duration) time.Duration {
	if len(durs) == 0 {
		return 0
	}
	m := meanDur(durs)
	var varianceSum float64
	for _, d := range durs {
		diff := float64(d - m)
		varianceSum += diff * diff
	}
	return time.Duration(math.Sqrt(varianceSum / float64(len(durs))))
}

func meanF(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var sum float64
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func stddevF(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	m := meanF(vals)
	var varianceSum float64
	for _, v := range vals {
		diff := v - m
		varianceSum += diff * diff
	}
	return math.Sqrt(varianceSum / float64(len(vals)))
}
