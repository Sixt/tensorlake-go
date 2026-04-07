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
// launching N sandboxes concurrently and measuring the time from
// CreateSandbox to SandboxStatusRunning for each.
//
// Usage:
//
//	export TENSORLAKE_API_KEY=<your-api-key>
//	go run ./examples/sandbox-benchmark -n 100
//
// Flags:
//
//	-n         Number of concurrent sandboxes to create (default: 100)
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
	"sync"
	"syscall"
	"time"

	tensorlake "github.com/sixt/tensorlake-go"
)

type result struct {
	index     int
	sandboxID string
	createDur time.Duration // time to get CreateSandbox response
	readyDur  time.Duration // time from create to running (includes createDur)
	err       error
}

func main() {
	n := flag.Int("n", 100, "number of concurrent sandboxes")
	timeout := flag.Int64("timeout", 120, "sandbox timeout in seconds")
	pollInterval := flag.Duration("poll", 500*time.Millisecond, "poll interval for status checks")
	maxWait := flag.Duration("max-wait", 120*time.Second, "maximum wait time per sandbox")
	flag.Parse()

	apiKey := os.Getenv("TENSORLAKE_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "TENSORLAKE_API_KEY environment variable is required")
		os.Exit(1)
	}

	c := tensorlake.NewClient(tensorlake.WithAPIKey(apiKey))
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	fmt.Fprintf(os.Stderr, "Launching %d sandboxes concurrently...\n", *n)
	overallStart := time.Now()

	results := make([]result, *n)
	var wg sync.WaitGroup

	// Launch all sandboxes concurrently.
	for i := range *n {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = bench(ctx, c, idx, *timeout, *pollInterval, *maxWait)
		}(i)
	}

	// Wait for all to complete.
	wg.Wait()
	overallDur := time.Since(overallStart)

	// Clean up all sandboxes.
	fmt.Fprintf(os.Stderr, "\nCleaning up sandboxes...\n")
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
	fmt.Fprintf(os.Stderr, "Cleanup complete.\n\n")

	// Analyze results.
	var succeeded, failed int
	var createDurs, readyDurs []time.Duration

	for _, r := range results {
		if r.err != nil {
			failed++
			fmt.Fprintf(os.Stderr, "  [%3d] FAIL: %v\n", r.index, r.err)
		} else {
			succeeded++
			createDurs = append(createDurs, r.createDur)
			readyDurs = append(readyDurs, r.readyDur)
		}
	}

	fmt.Printf("=== Sandbox Benchmark Results ===\n")
	fmt.Printf("Concurrency:    %d\n", *n)
	fmt.Printf("Succeeded:      %d\n", succeeded)
	fmt.Printf("Failed:         %d\n", failed)
	fmt.Printf("Total duration: %s\n\n", overallDur.Round(time.Millisecond))

	if succeeded > 0 {
		slices.SortFunc(createDurs, func(a, b time.Duration) int { return int(a - b) })
		slices.SortFunc(readyDurs, func(a, b time.Duration) int { return int(a - b) })

		fmt.Printf("--- Create API Response Time (time to receive sandbox_id) ---\n")
		printStats(createDurs)

		fmt.Printf("\n--- Time to Running (create → running) ---\n")
		printStats(readyDurs)
	}
}

func bench(ctx context.Context, c *tensorlake.Client, idx int, timeout int64, pollInterval, maxWait time.Duration) result {
	r := result{index: idx}

	// Create sandbox.
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

	// Already running? (rare but possible)
	if resp.Status == tensorlake.SandboxStatusRunning {
		r.readyDur = r.createDur
		return r
	}

	// Poll until running.
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

	r.err = fmt.Errorf("timed out waiting for running (last poll at %s)", time.Since(start).Round(time.Millisecond))
	return r
}

func printStats(durs []time.Duration) {
	n := len(durs)
	if n == 0 {
		return
	}

	var sum time.Duration
	for _, d := range durs {
		sum += d
	}
	mean := sum / time.Duration(n)

	var varianceSum float64
	for _, d := range durs {
		diff := float64(d - mean)
		varianceSum += diff * diff
	}
	stddev := time.Duration(math.Sqrt(varianceSum / float64(n)))

	fmt.Printf("  Min:    %s\n", durs[0].Round(time.Millisecond))
	fmt.Printf("  Max:    %s\n", durs[n-1].Round(time.Millisecond))
	fmt.Printf("  Mean:   %s\n", mean.Round(time.Millisecond))
	fmt.Printf("  Stddev: %s\n", stddev.Round(time.Millisecond))
	fmt.Printf("  P50:    %s\n", percentile(durs, 0.50).Round(time.Millisecond))
	fmt.Printf("  P75:    %s\n", percentile(durs, 0.75).Round(time.Millisecond))
	fmt.Printf("  P90:    %s\n", percentile(durs, 0.90).Round(time.Millisecond))
	fmt.Printf("  P95:    %s\n", percentile(durs, 0.95).Round(time.Millisecond))
	fmt.Printf("  P99:    %s\n", percentile(durs, 0.99).Round(time.Millisecond))
	fmt.Printf("  P99.9:  %s\n", percentile(durs, 0.999).Round(time.Millisecond))
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
