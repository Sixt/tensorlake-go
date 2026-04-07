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
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestPTYFullLifecycle exercises the complete PTY session lifecycle within
// a sandbox, covering every PTY API endpoint:
//
//	1. Create sandbox (prerequisite)
//	2. CreatePTY: start a bash session → get session_id + token
//	3. ListPTY: verify session appears in list
//	4. GetPTY: verify session details (command, is_alive, dimensions)
//	5. ResizePTY: change terminal dimensions
//	6. GetPTY: verify dimensions changed
//	7. ConnectPTY: open WebSocket, send Ready, write command, read output
//	8. KillPTY: terminate session
//	9. GetPTY: verify session is no longer alive
//	10. Delete sandbox (cleanup)
//
// Prerequisites:
//   - TENSORLAKE_API_KEY env var must be set (test skips otherwise).
func TestPTYFullLifecycle(t *testing.T) {
	c := initializeSandboxClient(t)

	// ── Create sandbox ──────────────────────────────────────────
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

	// ── Step 1: Create PTY session ──────────────────────────────
	// The PTY daemon may take a moment to become ready after sandbox status
	// transitions to running. Retry a few times.
	t.Log("=== create PTY")
	var ptyResp *CreatePTYResponse
	for range 10 {
		ptyResp, err = c.CreatePTY(t.Context(), sandboxID, &CreatePTYRequest{
			Command:    "/bin/sh",
			Env:        map[string]string{"TERM": "xterm-256color"},
			WorkingDir: "/",
			Rows:       24,
			Cols:       80,
		})
		if err == nil {
			break
		}
		t.Logf("CreatePTY not ready yet: %v", err)
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		t.Fatalf("CreatePTY: %v", err)
	}
	if ptyResp.SessionId == "" {
		t.Fatal("SessionId is empty")
	}
	if ptyResp.Token == "" {
		t.Fatal("Token is empty")
	}
	sessionID := ptyResp.SessionId
	token := ptyResp.Token
	t.Logf("PTY session created: id=%s", sessionID)

	// Ensure cleanup.
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = c.KillPTY(ctx, sandboxID, sessionID)
	})

	// ── Step 2: List PTY sessions ───────────────────────────────
	t.Log("=== list PTY sessions")
	listResp, err := c.ListPTY(t.Context(), sandboxID)
	if err != nil {
		t.Fatalf("ListPTY: %v", err)
	}
	found := false
	for _, s := range listResp.Sessions {
		if s.SessionId == sessionID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("session %s not found in list (%d sessions)", sessionID, len(listResp.Sessions))
	}

	// ── Step 3: Get PTY session details ─────────────────────────
	t.Log("=== get PTY session")
	info, err := c.GetPTY(t.Context(), sandboxID, sessionID)
	if err != nil {
		t.Fatalf("GetPTY: %v", err)
	}
	if info.Command != "/bin/sh" {
		t.Errorf("Command = %q, want /bin/sh", info.Command)
	}
	if !info.IsAlive {
		t.Error("IsAlive = false, want true")
	}
	if info.Rows != 24 || info.Cols != 80 {
		t.Errorf("dimensions = %dx%d, want 24x80", info.Rows, info.Cols)
	}
	t.Logf("PTY info: pid=%d, command=%s, rows=%d, cols=%d", info.PID, info.Command, info.Rows, info.Cols)

	// ── Step 4: Resize PTY ──────────────────────────────────────
	t.Log("=== resize PTY")
	err = c.ResizePTY(t.Context(), sandboxID, sessionID, &ResizePTYRequest{
		Rows: 40,
		Cols: 120,
	})
	if err != nil {
		t.Fatalf("ResizePTY: %v", err)
	}

	// Verify resize took effect.
	info, err = c.GetPTY(t.Context(), sandboxID, sessionID)
	if err != nil {
		t.Fatalf("GetPTY after resize: %v", err)
	}
	if info.Rows != 40 || info.Cols != 120 {
		t.Errorf("dimensions after resize = %dx%d, want 40x120", info.Rows, info.Cols)
	}

	// ── Step 5: WebSocket connection ────────────────────────────
	// Connect, send Ready, write "echo hello\n", read output containing "hello".
	t.Log("=== connect PTY websocket")
	conn, err := c.ConnectPTY(t.Context(), sandboxID, sessionID, token)
	if err != nil {
		t.Fatalf("ConnectPTY: %v", err)
	}
	defer conn.Close()

	// Send Ready signal — required before receiving output.
	err = conn.Ready(t.Context())
	if err != nil {
		t.Fatalf("Ready: %v", err)
	}

	// Write a command.
	err = conn.Write(t.Context(), []byte("echo hello-from-pty-test\n"))
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	// Read output until we see "hello-from-pty-test" or timeout.
	var output strings.Builder
	readCtx, readCancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer readCancel()
	for {
		msg, err := conn.Read(readCtx)
		if err != nil {
			t.Fatalf("Read: %v (output so far: %q)", err, output.String())
		}
		if msg.Type == PTYMessageData {
			output.Write(msg.Data)
			if strings.Contains(output.String(), "hello-from-pty-test") {
				break
			}
		}
		if msg.Type == PTYMessageExit {
			t.Fatalf("unexpected exit with code %d", msg.ExitCode)
		}
	}
	t.Logf("PTY output contains expected string")

	// Test resize via WebSocket.
	err = conn.Resize(t.Context(), 100, 30)
	if err != nil {
		t.Fatalf("Resize via WebSocket: %v", err)
	}

	conn.Close()

	// ── Step 6: Kill PTY session ────────────────────────────────
	t.Log("=== kill PTY")
	err = c.KillPTY(t.Context(), sandboxID, sessionID)
	if err != nil {
		t.Fatalf("KillPTY: %v", err)
	}

	// Wait briefly for kill to take effect.
	time.Sleep(1 * time.Second)

	// Verify session is no longer alive.
	// The kill sends SIGHUP then SIGKILL; exit_code may or may not be set
	// depending on how quickly the process terminates.
	info, err = c.GetPTY(t.Context(), sandboxID, sessionID)
	if err != nil {
		t.Fatalf("GetPTY after kill: %v", err)
	}
	if info.IsAlive {
		t.Error("IsAlive = true after kill, want false")
	}
	t.Logf("PTY killed: is_alive=%v, exit_code=%v", info.IsAlive, info.ExitCode)

	t.Log("=== PTY lifecycle complete")
}

// TestPTYSessionInfoDeserialization verifies JSON deserialization of
// PTYSessionInfo, including nullable fields (ended_at, exit_code).
func TestPTYSessionInfoDeserialization(t *testing.T) {
	raw := `{
		"session_id": "abc123",
		"pid": 42,
		"command": "/bin/bash",
		"args": ["-l"],
		"rows": 24,
		"cols": 80,
		"created_at": 1704067200000,
		"ended_at": null,
		"exit_code": null,
		"is_alive": true
	}`

	var info PTYSessionInfo
	if err := json.Unmarshal([]byte(raw), &info); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if info.SessionId != "abc123" {
		t.Errorf("SessionId = %q, want %q", info.SessionId, "abc123")
	}
	if info.PID != 42 {
		t.Errorf("PID = %d, want 42", info.PID)
	}
	if !info.IsAlive {
		t.Error("IsAlive = false, want true")
	}
	if info.EndedAt != nil {
		t.Errorf("EndedAt = %v, want nil", info.EndedAt)
	}
	if info.ExitCode != nil {
		t.Errorf("ExitCode = %v, want nil", info.ExitCode)
	}

	// Test with ended session.
	raw2 := `{
		"session_id": "def456",
		"pid": 100,
		"command": "/bin/sh",
		"args": [],
		"rows": 40,
		"cols": 120,
		"created_at": 1704067200000,
		"ended_at": 1704070800000,
		"exit_code": 0,
		"is_alive": false
	}`

	var info2 PTYSessionInfo
	if err := json.Unmarshal([]byte(raw2), &info2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if info2.IsAlive {
		t.Error("IsAlive = true, want false")
	}
	if info2.EndedAt == nil || *info2.EndedAt != 1704070800000 {
		t.Errorf("EndedAt = %v, want 1704070800000", info2.EndedAt)
	}
	if info2.ExitCode == nil || *info2.ExitCode != 0 {
		t.Errorf("ExitCode = %v, want 0", info2.ExitCode)
	}
}
