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

func TestSandboxLifecycle(t *testing.T) {
	c := initializeSandboxClient(t)

	// Create sandbox.
	createResp, err := c.CreateSandbox(t.Context(), &CreateSandboxRequest{
		TimeoutSecs: ptr(int64(300)),
	})
	if err != nil {
		t.Fatalf("failed to create sandbox: %v", err)
	}
	t.Logf("sandbox created: %+v", createResp)

	if createResp.SandboxId == "" {
		t.Fatal("SandboxId is empty")
	}

	sandboxID := createResp.SandboxId
	t.Cleanup(func() {
		_ = c.DeleteSandbox(t.Context(), sandboxID)
	})

	// List sandboxes.
	listResp, err := c.ListSandboxes(t.Context(), &ListSandboxesRequest{Limit: 10})
	if err != nil {
		t.Fatalf("failed to list sandboxes: %v", err)
	}
	t.Logf("listed %d sandboxes", len(listResp.Sandboxes))

	found := false
	for _, sb := range listResp.Sandboxes {
		if sb.Id == sandboxID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("sandbox %s not found in list", sandboxID)
	}

	// Get sandbox.
	info, err := c.GetSandbox(t.Context(), sandboxID)
	if err != nil {
		t.Fatalf("failed to get sandbox: %v", err)
	}
	t.Logf("sandbox info: %+v", info)

	if info.Id != sandboxID {
		t.Errorf("Id = %q, want %q", info.Id, sandboxID)
	}

	// Update sandbox.
	updated, err := c.UpdateSandbox(t.Context(), sandboxID, &UpdateSandboxRequest{
		ExposedPorts: []int32{8080},
	})
	if err != nil {
		t.Fatalf("failed to update sandbox: %v", err)
	}
	t.Logf("sandbox updated: %+v", updated)

	// Delete sandbox.
	err = c.DeleteSandbox(t.Context(), sandboxID)
	if err != nil {
		t.Fatalf("failed to delete sandbox: %v", err)
	}
	t.Log("sandbox deleted")
}

func TestSandboxSuspendResume(t *testing.T) {
	c := initializeSandboxClient(t)

	// Create a named sandbox (only named sandboxes can be suspended).
	name := fmt.Sprintf("test-sr-%d", time.Now().UnixNano())
	createResp, err := c.CreateSandbox(t.Context(), &CreateSandboxRequest{
		Name:        name,
		TimeoutSecs: ptr(int64(300)),
	})
	if err != nil {
		t.Fatalf("failed to create sandbox: %v", err)
	}
	sandboxID := createResp.SandboxId
	t.Logf("sandbox created: %s", sandboxID)

	t.Cleanup(func() {
		_ = c.ResumeSandbox(t.Context(), sandboxID)
		_ = c.DeleteSandbox(t.Context(), sandboxID)
	})

	// Wait for sandbox to be running before suspend.
	info, err := c.GetSandbox(t.Context(), sandboxID)
	if err != nil {
		t.Fatalf("failed to get sandbox: %v", err)
	}
	t.Logf("sandbox status: %s", info.Status)

	// Suspend.
	err = c.SuspendSandbox(t.Context(), sandboxID)
	if err != nil {
		t.Fatalf("failed to suspend sandbox: %v", err)
	}
	t.Log("sandbox suspend initiated")

	// Wait for sandbox to be suspended.
	for range 30 {
		info, err = c.GetSandbox(t.Context(), sandboxID)
		if err != nil {
			t.Fatalf("failed to get sandbox: %v", err)
		}
		if info.Status == SandboxStatusSuspended {
			break
		}
		t.Logf("sandbox status: %s, waiting...", info.Status)
		time.Sleep(2 * time.Second)
	}
	if info.Status != SandboxStatusSuspended {
		t.Fatalf("sandbox did not suspend, status: %s", info.Status)
	}

	// Resume.
	err = c.ResumeSandbox(t.Context(), sandboxID)
	if err != nil {
		t.Fatalf("failed to resume sandbox: %v", err)
	}
	t.Log("sandbox resume initiated")
}

func TestSandboxSnapshot(t *testing.T) {
	c := initializeSandboxClient(t)

	createResp, err := c.CreateSandbox(t.Context(), &CreateSandboxRequest{
		TimeoutSecs: ptr(int64(300)),
	})
	if err != nil {
		t.Fatalf("failed to create sandbox: %v", err)
	}
	sandboxID := createResp.SandboxId
	t.Logf("sandbox created: %s", sandboxID)

	t.Cleanup(func() {
		_ = c.DeleteSandbox(t.Context(), sandboxID)
	})

	// Snapshot.
	snapResp, err := c.SnapshotSandbox(t.Context(), sandboxID, nil)
	if err != nil {
		t.Fatalf("failed to snapshot sandbox: %v", err)
	}
	t.Logf("snapshot created: %+v", snapResp)

	if snapResp.SnapshotId == "" {
		t.Error("SnapshotId is empty")
	}
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
