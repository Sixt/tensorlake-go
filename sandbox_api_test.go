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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

func initializeSandboxClient(t *testing.T) *Client {
	t.Helper()
	apiKey := os.Getenv("TENSORLAKE_API_KEY")
	if apiKey == "" {
		t.Skip("TENSORLAKE_API_KEY must be set")
	}
	return NewClient(WithAPIKey(apiKey))
}

// waitForStatus polls GetSandbox until the sandbox reaches the desired status
// or the timeout (60s) is exceeded.
func waitForStatus(t *testing.T, c *Client, sandboxID string, want SandboxStatus) *SandboxInfo {
	t.Helper()
	var info *SandboxInfo
	var err error
	for range 30 {
		info, err = c.GetSandbox(t.Context(), sandboxID)
		if err != nil {
			t.Fatalf("failed to get sandbox %s: %v", sandboxID, err)
		}
		if info.Status == want {
			return info
		}
		t.Logf("sandbox %s: status=%s, waiting for %s...", sandboxID, info.Status, want)
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("sandbox %s: timed out waiting for status %s (current: %s)", sandboxID, want, info.Status)
	return nil
}

// deleteSandbox is a cleanup helper that terminates a sandbox, ignoring errors.
// Uses a fresh background context to avoid test context cancellation.
func deleteSandbox(t *testing.T, c *Client, sandboxID string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := c.DeleteSandbox(ctx, sandboxID); err != nil {
		t.Logf("cleanup: failed to delete sandbox %s: %v", sandboxID, err)
	}
}

// TestSandboxFullLifecycle exercises every state transition:
//
//	create → pending → running
//	  → file write/read/list/delete (file operations while running)
//	  → update (change exposed ports)
//	  → snapshot (running → snapshotting → running)
//	  → suspend (running → suspending → suspended)
//	  → resume (suspended → pending → running)
//	  → delete (running → terminated)
//	  → restore from snapshot (create with snapshot_id)
//	  → delete restored sandbox
func TestSandboxFullLifecycle(t *testing.T) {
	c := initializeSandboxClient(t)

	name := fmt.Sprintf("test-lifecycle-%d", time.Now().UnixNano())

	// ── Create ──────────────────────────────────────────────────
	t.Log("=== create sandbox")
	createResp, err := c.CreateSandbox(t.Context(), &CreateSandboxRequest{
		Name:        name,
		TimeoutSecs: ptr(int64(600)),
	})
	if err != nil {
		t.Fatalf("CreateSandbox: %v", err)
	}
	sandboxID := createResp.SandboxId
	t.Logf("created sandbox %s, status=%s", sandboxID, createResp.Status)

	// Register cleanup first — always runs, even on fatal.
	t.Cleanup(func() { deleteSandbox(t, c, sandboxID) })

	if createResp.Status != SandboxStatusPending && createResp.Status != SandboxStatusRunning {
		t.Fatalf("unexpected initial status: %s", createResp.Status)
	}

	// ── pending → running ───────────────────────────────────────
	t.Log("=== wait for running")
	info := waitForStatus(t, c, sandboxID, SandboxStatusRunning)
	t.Logf("sandbox running: cpus=%.1f memory_mb=%d", info.Resources.CPUs, info.Resources.MemoryMB)

	// ── List sandboxes ──────────────────────────────────────────
	t.Log("=== list sandboxes")
	listResp, err := c.ListSandboxes(t.Context(), &ListSandboxesRequest{Limit: 50})
	if err != nil {
		t.Fatalf("ListSandboxes: %v", err)
	}
	found := false
	for _, sb := range listResp.Sandboxes {
		if sb.Id == sandboxID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("sandbox %s not found in list (%d sandboxes)", sandboxID, len(listResp.Sandboxes))
	}

	// ── Get sandbox ─────────────────────────────────────────────
	t.Log("=== get sandbox")
	info, err = c.GetSandbox(t.Context(), sandboxID)
	if err != nil {
		t.Fatalf("GetSandbox: %v", err)
	}
	if info.Name != name {
		t.Errorf("Name = %q, want %q", info.Name, name)
	}

	// ── Update sandbox ──────────────────────────────────────────
	t.Log("=== update sandbox")
	updated, err := c.UpdateSandbox(t.Context(), sandboxID, &UpdateSandboxRequest{
		ExposedPorts: []int32{8080, 3000},
	})
	if err != nil {
		t.Fatalf("UpdateSandbox: %v", err)
	}
	if len(updated.ExposedPorts) != 2 {
		t.Errorf("ExposedPorts = %v, want [8080, 3000]", updated.ExposedPorts)
	}

	// ── File operations (while running) ─────────────────────────
	t.Log("=== file operations")
	content := []byte("lifecycle test content")
	err = c.WriteSandboxFile(t.Context(), sandboxID, "/workspace/lifecycle.txt", bytes.NewReader(content))
	if err != nil {
		t.Fatalf("WriteSandboxFile: %v", err)
	}

	data, err := c.ReadSandboxFile(t.Context(), sandboxID, "/workspace/lifecycle.txt")
	if err != nil {
		t.Fatalf("ReadSandboxFile: %v", err)
	}
	if string(data) != string(content) {
		t.Errorf("file content = %q, want %q", string(data), string(content))
	}

	dirResp, err := c.ListSandboxDirectory(t.Context(), sandboxID, "/workspace")
	if err != nil {
		t.Fatalf("ListSandboxDirectory: %v", err)
	}
	foundFile := false
	for _, entry := range dirResp.Entries {
		if entry.Name == "lifecycle.txt" && !entry.IsDir {
			foundFile = true
			break
		}
	}
	if !foundFile {
		t.Error("lifecycle.txt not found in directory listing")
	}

	err = c.DeleteSandboxFile(t.Context(), sandboxID, "/workspace/lifecycle.txt")
	if err != nil {
		t.Fatalf("DeleteSandboxFile: %v", err)
	}

	_, err = c.ReadSandboxFile(t.Context(), sandboxID, "/workspace/lifecycle.txt")
	if err == nil {
		t.Error("expected error reading deleted file, got nil")
	}

	// ── Snapshot (running → snapshotting → running) ─────────────
	t.Log("=== snapshot sandbox")
	snapResp, err := c.SnapshotSandbox(t.Context(), sandboxID, nil)
	if err != nil {
		t.Fatalf("SnapshotSandbox: %v", err)
	}
	if snapResp.SnapshotId == "" {
		t.Fatal("SnapshotId is empty")
	}
	snapshotID := snapResp.SnapshotId
	t.Logf("snapshot created: %s (status=%s)", snapshotID, snapResp.Status)

	// Wait for snapshot to complete. The sandbox status may briefly show
	// "snapshotting" then return to "running", but the snapshot operation
	// can still be in progress internally. Poll until suspend succeeds.
	t.Log("waiting for snapshot to complete...")
	for range 30 {
		err = c.SuspendSandbox(t.Context(), sandboxID)
		if err == nil {
			break
		}
		t.Logf("suspend not ready yet: %v", err)
		time.Sleep(2 * time.Second)
	}

	// ── Suspend (running → suspending → suspended) ──────────────
	t.Log("=== suspend sandbox")
	if err != nil {
		t.Fatalf("SuspendSandbox: %v", err)
	}
	waitForStatus(t, c, sandboxID, SandboxStatusSuspended)
	t.Log("sandbox suspended")

	// ── Resume (suspended → pending → running) ──────────────────
	t.Log("=== resume sandbox")
	err = c.ResumeSandbox(t.Context(), sandboxID)
	if err != nil {
		t.Fatalf("ResumeSandbox: %v", err)
	}
	waitForStatus(t, c, sandboxID, SandboxStatusRunning)
	t.Log("sandbox resumed and running")

	// ── Delete (running → terminated) ───────────────────────────
	t.Log("=== delete sandbox")
	err = c.DeleteSandbox(t.Context(), sandboxID)
	if err != nil {
		t.Fatalf("DeleteSandbox: %v", err)
	}

	// Verify terminated.
	info, err = c.GetSandbox(t.Context(), sandboxID)
	if err != nil {
		t.Fatalf("GetSandbox after delete: %v", err)
	}
	if info.Status != SandboxStatusTerminated {
		t.Errorf("status after delete = %s, want terminated", info.Status)
	}
	t.Log("sandbox terminated")

	// ── Idempotent delete ───────────────────────────────────────
	t.Log("=== idempotent delete")
	err = c.DeleteSandbox(t.Context(), sandboxID)
	if err != nil {
		t.Fatalf("idempotent DeleteSandbox: %v", err)
	}

	// ── Restore from snapshot ───────────────────────────────────
	t.Log("=== restore from snapshot")
	restoreResp, err := c.CreateSandbox(t.Context(), &CreateSandboxRequest{
		SnapshotId:  snapshotID,
		TimeoutSecs: ptr(int64(300)),
	})
	if err != nil {
		t.Fatalf("CreateSandbox (restore): %v", err)
	}
	restoredID := restoreResp.SandboxId
	t.Logf("restored sandbox %s from snapshot %s", restoredID, snapshotID)

	// Clean up restored sandbox.
	t.Cleanup(func() { deleteSandbox(t, c, restoredID) })

	waitForStatus(t, c, restoredID, SandboxStatusRunning)
	t.Log("restored sandbox running")

	err = c.DeleteSandbox(t.Context(), restoredID)
	if err != nil {
		t.Fatalf("DeleteSandbox (restored): %v", err)
	}
	t.Log("restored sandbox terminated")

	t.Log("=== full lifecycle complete")
}

func TestSandboxInfoDeserialization(t *testing.T) {
	raw := `{
		"id": "sb_abc",
		"namespace": "ns_def",
		"image": "custom:latest",
		"status": "running",
		"created_at": 1704067200000,
		"container_id": "ctr_123",
		"executor_id": "exec_456",
		"resources": {"cpus": 4.0, "memory_mb": 8192, "ephemeral_disk_mb": 20480},
		"timeout_secs": 7200,
		"sandbox_url": "https://sb_abc.sandbox.tensorlake.ai",
		"pool_id": "pool_1",
		"network_policy": {"allow_internet_access": false, "allow_out": ["10.0.0.0/8"], "deny_out": ["192.168.0.0/16"]},
		"allow_unauthenticated_access": true,
		"exposed_ports": [8080, 3000],
		"template_id": "tmpl_1",
		"name": "my-sandbox"
	}`

	var info SandboxInfo
	if err := json.Unmarshal([]byte(raw), &info); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if info.Id != "sb_abc" {
		t.Errorf("Id = %q, want %q", info.Id, "sb_abc")
	}
	if info.Status != SandboxStatusRunning {
		t.Errorf("Status = %q, want %q", info.Status, SandboxStatusRunning)
	}
	if info.Resources.CPUs != 4.0 {
		t.Errorf("CPUs = %f, want 4.0", info.Resources.CPUs)
	}
	if info.NetworkPolicy == nil {
		t.Fatal("NetworkPolicy is nil")
	}
	if info.NetworkPolicy.AllowInternetAccess {
		t.Error("AllowInternetAccess = true, want false")
	}
	if len(info.NetworkPolicy.AllowOut) != 1 || info.NetworkPolicy.AllowOut[0] != "10.0.0.0/8" {
		t.Errorf("AllowOut = %v, want [10.0.0.0/8]", info.NetworkPolicy.AllowOut)
	}
	if len(info.ExposedPorts) != 2 {
		t.Errorf("ExposedPorts length = %d, want 2", len(info.ExposedPorts))
	}
	if info.Name != "my-sandbox" {
		t.Errorf("Name = %q, want %q", info.Name, "my-sandbox")
	}
}

func ptr[T any](v T) *T { return &v }
