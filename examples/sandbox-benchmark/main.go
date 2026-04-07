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
// Usage:
//
//	export TENSORLAKE_API_KEY=<your-api-key>
//	go run ./examples/sandbox-benchmark 1 10 100
//	go run ./examples/sandbox-benchmark 1 10 100 1000
//
// Each positional argument is a concurrency level. They run sequentially,
// one after the other. All sandboxes from a level are cleaned up before
// the next level starts.
//
// Flags:
//
//	-timeout   Sandbox timeout in seconds (default: 120)
//	-poll      Poll interval for status checks (default: 500ms)
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
	"sync"
	"syscall"
	"time"

	tensorlake "github.com/sixt/tensorlake-go"
)

type result struct {
	index     int
	sandboxID string
	createDur time.Duration
	readyDur  time.Duration
	err       error
}

type levelStats struct {
	concurrency int
	succeeded   int
	failed      int
	totalDur    time.Duration
	createDurs  []time.Duration
	readyDurs   []time.Duration
}

func main() {
	timeout := flag.Int64("timeout", 120, "sandbox timeout in seconds")
	pollInterval := flag.Duration("poll", 500*time.Millisecond, "poll interval for status checks")
	maxWait := flag.Duration("max-wait", 120*time.Second, "maximum wait time per sandbox")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		args = []string{"1", "10", "100"}
	}

	var levels []int
	for _, arg := range args {
		n, err := strconv.Atoi(arg)
		if err != nil || n <= 0 {
			fmt.Fprintf(os.Stderr, "invalid concurrency level: %q\n", arg)
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

	var allStats []levelStats

	for i, n := range levels {
		if ctx.Err() != nil {
			break
		}
		if i > 0 {
			fmt.Fprintf(os.Stderr, "\n")
		}
		fmt.Fprintf(os.Stderr, "=== Level %d/%d: %d concurrent sandboxes ===\n", i+1, len(levels), n)
		stats := runLevel(ctx, c, n, *timeout, *pollInterval, *maxWait)
		allStats = append(allStats, stats)
	}

	// Print summary table.
	fmt.Printf("\n")
	printSummary(allStats)
}

func runLevel(ctx context.Context, c *tensorlake.Client, n int, timeout int64, pollInterval, maxWait time.Duration) levelStats {
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

	// Collect stats.
	stats := levelStats{concurrency: n, totalDur: totalDur}
	for _, r := range results {
		if r.err != nil {
			stats.failed++
			fmt.Fprintf(os.Stderr, "  [%3d] FAIL: %v\n", r.index, r.err)
		} else {
			stats.succeeded++
			stats.createDurs = append(stats.createDurs, r.createDur)
			stats.readyDurs = append(stats.readyDurs, r.readyDur)
		}
	}

	slices.SortFunc(stats.createDurs, func(a, b time.Duration) int { return int(a - b) })
	slices.SortFunc(stats.readyDurs, func(a, b time.Duration) int { return int(a - b) })

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
		r.readyDur = r.createDur
		return r
	}

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

func printSummary(allStats []levelStats) {
	pcts := []struct {
		name string
		p    float64
	}{
		{"P50", 0.50}, {"P75", 0.75}, {"P90", 0.90},
		{"P95", 0.95}, {"P99", 0.99}, {"P99.9", 0.999},
	}

	// Header.
	fmt.Printf("%-12s", "Concurrency")
	for _, s := range allStats {
		fmt.Printf("  %10d", s.concurrency)
	}
	fmt.Println()
	fmt.Printf("%-12s", "Succeeded")
	for _, s := range allStats {
		fmt.Printf("  %10d", s.succeeded)
	}
	fmt.Println()
	fmt.Printf("%-12s", "Failed")
	for _, s := range allStats {
		fmt.Printf("  %10d", s.failed)
	}
	fmt.Println()
	fmt.Printf("%-12s", "Wall time")
	for _, s := range allStats {
		fmt.Printf("  %10s", s.totalDur.Round(time.Millisecond))
	}
	fmt.Println()

	// Create API stats.
	fmt.Printf("\n--- Create API Response Time ---\n")
	printMetricRows(allStats, pcts, func(s levelStats) []time.Duration { return s.createDurs })

	// Time to running stats.
	fmt.Printf("\n--- Time to Running ---\n")
	printMetricRows(allStats, pcts, func(s levelStats) []time.Duration { return s.readyDurs })
}

func printMetricRows(allStats []levelStats, pcts []struct {
	name string
	p    float64
}, getDurs func(levelStats) []time.Duration) {
	// Min.
	fmt.Printf("%-12s", "Min")
	for _, s := range allStats {
		d := getDurs(s)
		if len(d) > 0 {
			fmt.Printf("  %10s", d[0].Round(time.Millisecond))
		} else {
			fmt.Printf("  %10s", "-")
		}
	}
	fmt.Println()

	// Mean.
	fmt.Printf("%-12s", "Mean")
	for _, s := range allStats {
		d := getDurs(s)
		if len(d) > 0 {
			fmt.Printf("  %10s", mean(d).Round(time.Millisecond))
		} else {
			fmt.Printf("  %10s", "-")
		}
	}
	fmt.Println()

	// Stddev.
	fmt.Printf("%-12s", "Stddev")
	for _, s := range allStats {
		d := getDurs(s)
		if len(d) > 0 {
			fmt.Printf("  %10s", stddev(d).Round(time.Millisecond))
		} else {
			fmt.Printf("  %10s", "-")
		}
	}
	fmt.Println()

	// Percentiles.
	for _, p := range pcts {
		fmt.Printf("%-12s", p.name)
		for _, s := range allStats {
			d := getDurs(s)
			if len(d) > 0 {
				fmt.Printf("  %10s", percentile(d, p.p).Round(time.Millisecond))
			} else {
				fmt.Printf("  %10s", "-")
			}
		}
		fmt.Println()
	}

	// Max.
	fmt.Printf("%-12s", "Max")
	for _, s := range allStats {
		d := getDurs(s)
		if len(d) > 0 {
			fmt.Printf("  %10s", d[len(d)-1].Round(time.Millisecond))
		} else {
			fmt.Printf("  %10s", "-")
		}
	}
	fmt.Println()
}

func mean(durs []time.Duration) time.Duration {
	if len(durs) == 0 {
		return 0
	}
	var sum time.Duration
	for _, d := range durs {
		sum += d
	}
	return sum / time.Duration(len(durs))
}

func stddev(durs []time.Duration) time.Duration {
	if len(durs) == 0 {
		return 0
	}
	m := mean(durs)
	var varianceSum float64
	for _, d := range durs {
		diff := float64(d - m)
		varianceSum += diff * diff
	}
	return time.Duration(math.Sqrt(varianceSum / float64(len(durs))))
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
