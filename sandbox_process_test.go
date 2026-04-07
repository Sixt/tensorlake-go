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

package tensorlake

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestProcessFullLifecycle exercises the complete process API lifecycle
// within a sandbox:
//
//  1. Create sandbox and wait for running
//  2. StartProcess: run "echo hello" with stdout capture
//  3. ListProcesses: verify process appears
//  4. GetProcess: verify process details
//  5. GetProcessStdout: verify captured output contains "hello"
//  6. GetProcessOutput: verify merged output
//  7. StartProcess with stdin pipe: run "cat"
//  8. WriteProcessStdin: write data to cat's stdin
//  9. CloseProcessStdin: close stdin (cat exits)
//  10. GetProcessStdout: verify cat echoed the data
//  11. StartProcess: run "sleep 60" for signal/kill testing
//  12. SignalProcess: send SIGTERM (15)
//  13. GetProcess: verify process was signaled
//  14. StartProcess: run another "sleep 60"
//  15. KillProcess: kill the process
//  16. FollowProcessStdout: start a process, follow stdout via SSE
//  17. FollowProcessOutput: follow merged output via SSE
//  18. Cleanup: delete sandbox
func TestProcessFullLifecycle(t *testing.T) {
	c := initializeSandboxClient(t)

	// Create sandbox.
	createResp, err := c.CreateSandbox(t.Context(), &CreateSandboxRequest{
		TimeoutSecs: ptr(int64(300)),
	})
	if err != nil {
		t.Fatalf("CreateSandbox: %v", err)
	}
	sandboxID := createResp.SandboxId
	t.Logf("sandbox created: %s", sandboxID)
	t.Cleanup(func() { deleteSandbox(t, c, sandboxID) })

	waitForStatus(t, c, sandboxID, SandboxStatusRunning)

	// ── Step 1: Start a simple process ──────────────────────────
	t.Log("=== start echo process")
	var echoProc *ProcessInfo
	for range 10 {
		echoProc, err = c.StartProcess(t.Context(), sandboxID, &StartProcessRequest{
			Command:    "/bin/sh",
			Args:       []string{"-c", "echo hello-from-process"},
			StdoutMode: OutputModeCapture,
		})
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		t.Fatalf("StartProcess: %v", err)
	}
	t.Logf("process started: pid=%d status=%s", echoProc.PID, echoProc.Status)

	// Wait for process to finish.
	time.Sleep(2 * time.Second)

	// ── Step 2: List processes ───────────────────────────────────
	t.Log("=== list processes")
	listResp, err := c.ListProcesses(t.Context(), sandboxID)
	if err != nil {
		t.Fatalf("ListProcesses: %v", err)
	}
	found := false
	for _, p := range listResp.Processes {
		if p.PID == echoProc.PID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("pid %d not found in process list", echoProc.PID)
	}

	// ── Step 3: Get process details ─────────────────────────────
	t.Log("=== get process")
	info, err := c.GetProcess(t.Context(), sandboxID, echoProc.PID)
	if err != nil {
		t.Fatalf("GetProcess: %v", err)
	}
	if info.Command != "/bin/sh" {
		t.Errorf("Command = %q, want /bin/sh", info.Command)
	}
	t.Logf("process info: pid=%d status=%s exit_code=%v", info.PID, info.Status, info.ExitCode)

	// ── Step 4: Get stdout ──────────────────────────────────────
	t.Log("=== get stdout")
	stdout, err := c.GetProcessStdout(t.Context(), sandboxID, echoProc.PID)
	if err != nil {
		t.Fatalf("GetProcessStdout: %v", err)
	}
	if !containsLine(stdout.Lines, "hello-from-process") {
		t.Errorf("stdout lines = %v, want to contain 'hello-from-process'", stdout.Lines)
	}

	// ── Step 5: Get merged output ───────────────────────────────
	t.Log("=== get output")
	output, err := c.GetProcessOutput(t.Context(), sandboxID, echoProc.PID)
	if err != nil {
		t.Fatalf("GetProcessOutput: %v", err)
	}
	if output.LineCount == 0 {
		t.Error("output line_count = 0, want > 0")
	}

	// ── Step 5b: Get stderr ─────────────────────────────────────
	t.Log("=== get stderr")
	errProc, err := c.StartProcess(t.Context(), sandboxID, &StartProcessRequest{
		Command:    "/bin/sh",
		Args:       []string{"-c", "echo stderr-line >&2"},
		StderrMode: OutputModeCapture,
	})
	if err != nil {
		t.Fatalf("StartProcess (stderr): %v", err)
	}
	time.Sleep(2 * time.Second)

	stderrResp, err := c.GetProcessStderr(t.Context(), sandboxID, errProc.PID)
	if err != nil {
		t.Fatalf("GetProcessStderr: %v", err)
	}
	if !containsLine(stderrResp.Lines, "stderr-line") {
		t.Errorf("stderr lines = %v, want to contain 'stderr-line'", stderrResp.Lines)
	}

	// ── Step 6: Stdin pipe (cat) ────────────────────────────────
	t.Log("=== start cat with stdin pipe")
	catProc, err := c.StartProcess(t.Context(), sandboxID, &StartProcessRequest{
		Command:    "/bin/cat",
		StdinMode:  StdinModePipe,
		StdoutMode: OutputModeCapture,
	})
	if err != nil {
		t.Fatalf("StartProcess (cat): %v", err)
	}
	t.Logf("cat started: pid=%d stdin_writable=%v", catProc.PID, catProc.StdinWritable)

	if !catProc.StdinWritable {
		t.Error("StdinWritable = false, want true")
	}

	// Write to stdin.
	err = c.WriteProcessStdin(t.Context(), sandboxID, catProc.PID, bytes.NewReader([]byte("piped-data\n")))
	if err != nil {
		t.Fatalf("WriteProcessStdin: %v", err)
	}

	// Close stdin (cat will exit after reading EOF).
	err = c.CloseProcessStdin(t.Context(), sandboxID, catProc.PID)
	if err != nil {
		t.Fatalf("CloseProcessStdin: %v", err)
	}

	time.Sleep(2 * time.Second)

	// Verify cat's stdout.
	catStdout, err := c.GetProcessStdout(t.Context(), sandboxID, catProc.PID)
	if err != nil {
		t.Fatalf("GetProcessStdout (cat): %v", err)
	}
	if !containsLine(catStdout.Lines, "piped-data") {
		t.Errorf("cat stdout = %v, want to contain 'piped-data'", catStdout.Lines)
	}

	// ── Step 7: Signal process ──────────────────────────────────
	t.Log("=== signal process (SIGTERM)")
	sleepProc, err := c.StartProcess(t.Context(), sandboxID, &StartProcessRequest{
		Command: "/bin/sleep",
		Args:    []string{"60"},
	})
	if err != nil {
		t.Fatalf("StartProcess (sleep): %v", err)
	}

	err = c.SignalProcess(t.Context(), sandboxID, sleepProc.PID, &SignalProcessRequest{Signal: 15})
	if err != nil {
		t.Fatalf("SignalProcess: %v", err)
	}
	time.Sleep(time.Second)

	sigInfo, err := c.GetProcess(t.Context(), sandboxID, sleepProc.PID)
	if err != nil {
		t.Fatalf("GetProcess after signal: %v", err)
	}
	if sigInfo.Status == ProcessStatusRunning {
		t.Error("process still running after SIGTERM")
	}
	t.Logf("process after signal: status=%s signal=%v", sigInfo.Status, sigInfo.Signal)

	// ── Step 8: Kill process ────────────────────────────────────
	t.Log("=== kill process")
	sleep2, err := c.StartProcess(t.Context(), sandboxID, &StartProcessRequest{
		Command: "/bin/sleep",
		Args:    []string{"60"},
	})
	if err != nil {
		t.Fatalf("StartProcess (sleep2): %v", err)
	}

	err = c.KillProcess(t.Context(), sandboxID, sleep2.PID)
	if err != nil {
		t.Fatalf("KillProcess: %v", err)
	}
	time.Sleep(time.Second)

	killInfo, err := c.GetProcess(t.Context(), sandboxID, sleep2.PID)
	if err != nil {
		t.Fatalf("GetProcess after kill: %v", err)
	}
	if killInfo.Status == ProcessStatusRunning {
		t.Error("process still running after kill")
	}
	t.Logf("process after kill: status=%s", killInfo.Status)

	// ── Step 9: Follow stdout via SSE ───────────────────────────
	t.Log("=== follow stdout")
	seqProc, err := c.StartProcess(t.Context(), sandboxID, &StartProcessRequest{
		Command:    "/bin/sh",
		Args:       []string{"-c", "echo line1; echo line2; echo line3"},
		StdoutMode: OutputModeCapture,
	})
	if err != nil {
		t.Fatalf("StartProcess (seq): %v", err)
	}
	time.Sleep(2 * time.Second)

	var followLines []string
	for evt, err := range c.FollowProcessStdout(t.Context(), sandboxID, seqProc.PID) {
		if err != nil {
			t.Fatalf("FollowProcessStdout: %v", err)
		}
		followLines = append(followLines, evt.Line)
	}
	if !containsLine(followLines, "line1") || !containsLine(followLines, "line3") {
		t.Errorf("follow stdout lines = %v, want to contain line1 and line3", followLines)
	}
	t.Logf("follow stdout captured %d lines", len(followLines))

	// ── Step 9b: Follow stderr via SSE ──────────────────────────
	t.Log("=== follow stderr")
	errFollowProc, err := c.StartProcess(t.Context(), sandboxID, &StartProcessRequest{
		Command:    "/bin/sh",
		Args:       []string{"-c", "echo err1 >&2; echo err2 >&2"},
		StderrMode: OutputModeCapture,
	})
	if err != nil {
		t.Fatalf("StartProcess (follow stderr): %v", err)
	}
	time.Sleep(2 * time.Second)

	var followErrLines []string
	for evt, err := range c.FollowProcessStderr(t.Context(), sandboxID, errFollowProc.PID) {
		if err != nil {
			t.Fatalf("FollowProcessStderr: %v", err)
		}
		followErrLines = append(followErrLines, evt.Line)
	}
	if !containsLine(followErrLines, "err1") || !containsLine(followErrLines, "err2") {
		t.Errorf("follow stderr lines = %v, want to contain err1 and err2", followErrLines)
	}
	t.Logf("follow stderr captured %d lines", len(followErrLines))

	// ── Step 10: Follow merged output via SSE ───────────────────
	t.Log("=== follow output")
	mixProc, err := c.StartProcess(t.Context(), sandboxID, &StartProcessRequest{
		Command:    "/bin/sh",
		Args:       []string{"-c", "echo out-line; echo err-line >&2"},
		StdoutMode: OutputModeCapture,
		StderrMode: OutputModeCapture,
	})
	if err != nil {
		t.Fatalf("StartProcess (mix): %v", err)
	}
	time.Sleep(2 * time.Second)

	var outEvents []ProcessOutputEvent
	for evt, err := range c.FollowProcessOutput(t.Context(), sandboxID, mixProc.PID) {
		if err != nil {
			t.Fatalf("FollowProcessOutput: %v", err)
		}
		outEvents = append(outEvents, evt)
	}
	hasStdout := false
	hasStderr := false
	for _, evt := range outEvents {
		if evt.Stream == "stdout" {
			hasStdout = true
		}
		if evt.Stream == "stderr" {
			hasStderr = true
		}
	}
	if !hasStdout {
		t.Error("follow output missing stdout events")
	}
	if !hasStderr {
		t.Error("follow output missing stderr events")
	}
	t.Logf("follow output captured %d events (stdout=%v stderr=%v)", len(outEvents), hasStdout, hasStderr)

	t.Log("=== process lifecycle complete")
}

func TestProcessInfoDeserialization(t *testing.T) {
	raw := `{
		"pid": 42,
		"status": "exited",
		"exit_code": 0,
		"signal": null,
		"stdin_writable": false,
		"command": "python",
		"args": ["-m", "http.server"],
		"started_at": 1710000000000,
		"ended_at": 1710000060000
	}`

	var info ProcessInfo
	if err := json.Unmarshal([]byte(raw), &info); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if info.PID != 42 {
		t.Errorf("PID = %d, want 42", info.PID)
	}
	if info.Status != ProcessStatusExited {
		t.Errorf("Status = %q, want exited", info.Status)
	}
	if info.ExitCode == nil || *info.ExitCode != 0 {
		t.Errorf("ExitCode = %v, want 0", info.ExitCode)
	}
	if info.Signal != nil {
		t.Errorf("Signal = %v, want nil", info.Signal)
	}
	if info.EndedAt == nil || *info.EndedAt != 1710000060000 {
		t.Errorf("EndedAt = %v, want 1710000060000", info.EndedAt)
	}
}

func containsLine(lines []string, substr string) bool {
	for _, line := range lines {
		if strings.Contains(line, substr) {
			return true
		}
	}
	return false
}
