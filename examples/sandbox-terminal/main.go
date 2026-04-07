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

// Command sandbox-terminal creates a Tensorlake sandbox and connects an
// interactive terminal session to it via PTY over WebSocket.
//
// Usage:
//
//	export TENSORLAKE_API_KEY=<your-api-key>
//	go run ./examples/sandbox-terminal
//
// The program:
//  1. Creates a new sandbox (or reuses one via -sandbox flag)
//  2. Waits for the sandbox to reach "running" state
//  3. Creates a PTY session running /bin/sh
//  4. Connects via WebSocket and attaches stdin/stdout
//  5. Puts the local terminal in raw mode for interactive use
//  6. On exit (ctrl-d or process exit), cleans up the PTY session and sandbox
//
// Flags:
//
//	-sandbox   Reuse an existing sandbox ID instead of creating a new one
//	-timeout   Sandbox timeout in seconds (default: 300)
//	-keep      Do not delete the sandbox on exit
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/term"

	tensorlake "github.com/sixt/tensorlake-go"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	sandboxFlag := flag.String("sandbox", "", "reuse an existing sandbox ID")
	timeout := flag.Int64("timeout", 300, "sandbox timeout in seconds")
	keep := flag.Bool("keep", false, "do not delete the sandbox on exit")
	flag.Parse()

	apiKey := os.Getenv("TENSORLAKE_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("TENSORLAKE_API_KEY environment variable is required")
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("stdin is not a terminal; this program requires an interactive terminal")
	}

	c := tensorlake.NewClient(tensorlake.WithAPIKey(apiKey))
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Step 1: Create or reuse sandbox.
	sbID := *sandboxFlag
	if sbID == "" {
		fmt.Fprintf(os.Stderr, "Creating sandbox (timeout=%ds)...\n", *timeout)
		resp, err := c.CreateSandbox(ctx, &tensorlake.CreateSandboxRequest{
			TimeoutSecs: timeout,
		})
		if err != nil {
			return fmt.Errorf("create sandbox: %w", err)
		}
		sbID = resp.SandboxId
		fmt.Fprintf(os.Stderr, "Sandbox created: %s\n", sbID)

		if !*keep {
			defer func() {
				fmt.Fprintf(os.Stderr, "\nDeleting sandbox %s...\n", sbID)
				delCtx, delCancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer delCancel()
				if err := c.DeleteSandbox(delCtx, sbID); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to delete sandbox: %v\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "Sandbox deleted.\n")
				}
			}()
		}
	} else {
		fmt.Fprintf(os.Stderr, "Reusing sandbox: %s\n", sbID)
	}

	// Step 2: Wait for sandbox to be running.
	fmt.Fprintf(os.Stderr, "Waiting for sandbox to be running...")
	for range 60 {
		info, err := c.GetSandbox(ctx, sbID)
		if err != nil {
			return fmt.Errorf("get sandbox: %w", err)
		}
		if info.Status == tensorlake.SandboxStatusRunning {
			fmt.Fprintf(os.Stderr, " ready\n")
			break
		}
		if info.Status == tensorlake.SandboxStatusTerminated {
			return fmt.Errorf("sandbox is terminated")
		}
		fmt.Fprintf(os.Stderr, ".")
		time.Sleep(time.Second)
	}

	// Step 3: Get terminal size and create PTY session.
	cols, rows, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		cols, rows = 80, 24
	}

	fmt.Fprintf(os.Stderr, "Creating PTY session (%dx%d)...\n", cols, rows)
	var ptyResp *tensorlake.CreatePTYResponse
	for range 10 {
		ptyResp, err = c.CreatePTY(ctx, sbID, &tensorlake.CreatePTYRequest{
			Command: "/bin/sh",
			Env:     map[string]string{"TERM": os.Getenv("TERM")},
			Rows:    int32(rows),
			Cols:    int32(cols),
		})
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		return fmt.Errorf("create PTY: %w", err)
	}
	defer func() {
		killCtx, killCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer killCancel()
		_ = c.KillPTY(killCtx, sbID, ptyResp.SessionId)
	}()

	// Step 4: Connect WebSocket.
	fmt.Fprintf(os.Stderr, "Connecting to PTY session %s...\n", ptyResp.SessionId)
	conn, err := c.ConnectPTY(ctx, sbID, ptyResp.SessionId, ptyResp.Token)
	if err != nil {
		return fmt.Errorf("connect PTY: %w", err)
	}
	defer conn.Close()

	if err := conn.Ready(ctx); err != nil {
		return fmt.Errorf("send ready: %w", err)
	}

	// Step 5: Put local terminal in raw mode.
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("set raw mode: %w", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Handle terminal resize (SIGWINCH).
	winch := make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)
	go func() {
		for range winch {
			w, h, err := term.GetSize(int(os.Stdin.Fd()))
			if err == nil {
				_ = conn.Resize(ctx, uint16(w), uint16(h))
			}
		}
	}()

	done := make(chan struct{})

	// Read from PTY → stdout.
	go func() {
		defer close(done)
		for {
			msg, err := conn.Read(ctx)
			if err != nil {
				return
			}
			switch msg.Type {
			case tensorlake.PTYMessageData:
				os.Stdout.Write(msg.Data)
			case tensorlake.PTYMessageExit:
				return
			}
		}
	}()

	// Read from stdin → PTY.
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				return
			}
			if err := conn.Write(ctx, buf[:n]); err != nil {
				return
			}
		}
	}()

	// Wait for PTY to close or signal.
	select {
	case <-done:
	case <-ctx.Done():
	}

	return nil
}
