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
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestSandboxAPIServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	sandboxInfo := SandboxInfo{
		Id:                         "sb_123",
		Namespace:                  "ns_456",
		Status:                     SandboxStatusRunning,
		CreatedAt:                  1704067200000,
		Resources:                  ContainerResourcesInfo{CPUs: 2, MemoryMB: 4096, EphemeralDiskMB: 10240},
		TimeoutSecs:                3600,
		AllowUnauthenticatedAccess: false,
		Name:                       "test-sandbox",
	}

	// Create
	mux.HandleFunc("/sandboxes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			json.NewEncoder(w).Encode(CreateSandboxResponse{
				SandboxId: "sb_123",
				Status:    SandboxStatusPending,
			})
		case http.MethodGet:
			json.NewEncoder(w).Encode(ListSandboxesResponse{
				Sandboxes:  []SandboxInfo{sandboxInfo},
				NextCursor: "cursor_abc",
			})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Get / Update / Delete
	mux.HandleFunc("/sandboxes/sb_123", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(sandboxInfo)
		case http.MethodPatch:
			json.NewEncoder(w).Encode(sandboxInfo)
		case http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Not found
	mux.HandleFunc("/sandboxes/sb_missing", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("sandbox not found"))
	})

	// Snapshot
	mux.HandleFunc("/sandboxes/sb_123/snapshot", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(SnapshotSandboxResponse{
			SnapshotId: "snap_789",
			Status:     "creating",
		})
	})

	// Suspend
	mux.HandleFunc("/sandboxes/sb_123/suspend", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	// Resume
	mux.HandleFunc("/sandboxes/sb_123/resume", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	return httptest.NewServer(mux)
}

func newTestSandboxAPIClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	return NewClient(
		WithHTTPClient(srv.Client()),
		WithAPIKey("test-key"),
		WithSandboxAPIBaseURL(srv.URL),
	)
}

func TestCreateSandbox(t *testing.T) {
	srv := newTestSandboxAPIServer(t)
	defer srv.Close()
	c := newTestSandboxAPIClient(t, srv)

	resp, err := c.CreateSandbox(t.Context(), &CreateSandboxRequest{
		Name:        "test-sandbox",
		TimeoutSecs: ptr(int64(3600)),
	})
	if err != nil {
		t.Fatalf("CreateSandbox failed: %v", err)
	}
	if resp.SandboxId != "sb_123" {
		t.Errorf("SandboxId = %q, want %q", resp.SandboxId, "sb_123")
	}
	if resp.Status != SandboxStatusPending {
		t.Errorf("Status = %q, want %q", resp.Status, SandboxStatusPending)
	}
}

func TestCreateSandboxWithSnapshot(t *testing.T) {
	srv := newTestSandboxAPIServer(t)
	defer srv.Close()
	c := newTestSandboxAPIClient(t, srv)

	resp, err := c.CreateSandbox(t.Context(), &CreateSandboxRequest{
		SnapshotId: "snap_existing",
	})
	if err != nil {
		t.Fatalf("CreateSandbox (restore) failed: %v", err)
	}
	if resp.SandboxId != "sb_123" {
		t.Errorf("SandboxId = %q, want %q", resp.SandboxId, "sb_123")
	}
}

func TestListSandboxes(t *testing.T) {
	srv := newTestSandboxAPIServer(t)
	defer srv.Close()
	c := newTestSandboxAPIClient(t, srv)

	resp, err := c.ListSandboxes(t.Context(), &ListSandboxesRequest{Limit: 10})
	if err != nil {
		t.Fatalf("ListSandboxes failed: %v", err)
	}
	if len(resp.Sandboxes) != 1 {
		t.Fatalf("Sandboxes length = %d, want 1", len(resp.Sandboxes))
	}
	if resp.Sandboxes[0].Id != "sb_123" {
		t.Errorf("Id = %q, want %q", resp.Sandboxes[0].Id, "sb_123")
	}
	if resp.NextCursor != "cursor_abc" {
		t.Errorf("NextCursor = %q, want %q", resp.NextCursor, "cursor_abc")
	}
}

func TestGetSandbox(t *testing.T) {
	srv := newTestSandboxAPIServer(t)
	defer srv.Close()
	c := newTestSandboxAPIClient(t, srv)

	info, err := c.GetSandbox(t.Context(), "sb_123")
	if err != nil {
		t.Fatalf("GetSandbox failed: %v", err)
	}
	if info.Id != "sb_123" {
		t.Errorf("Id = %q, want %q", info.Id, "sb_123")
	}
	if info.Status != SandboxStatusRunning {
		t.Errorf("Status = %q, want %q", info.Status, SandboxStatusRunning)
	}
	if info.Name != "test-sandbox" {
		t.Errorf("Name = %q, want %q", info.Name, "test-sandbox")
	}
	if info.Resources.CPUs != 2 {
		t.Errorf("Resources.CPUs = %f, want 2", info.Resources.CPUs)
	}
}

func TestGetSandboxNotFound(t *testing.T) {
	srv := newTestSandboxAPIServer(t)
	defer srv.Close()
	c := newTestSandboxAPIClient(t, srv)

	_, err := c.GetSandbox(t.Context(), "sb_missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdateSandbox(t *testing.T) {
	srv := newTestSandboxAPIServer(t)
	defer srv.Close()
	c := newTestSandboxAPIClient(t, srv)

	info, err := c.UpdateSandbox(t.Context(), "sb_123", &UpdateSandboxRequest{
		AllowUnauthenticatedAccess: ptr(true),
		ExposedPorts:               []int32{8080, 3000},
	})
	if err != nil {
		t.Fatalf("UpdateSandbox failed: %v", err)
	}
	if info.Id != "sb_123" {
		t.Errorf("Id = %q, want %q", info.Id, "sb_123")
	}
}

func TestDeleteSandbox(t *testing.T) {
	srv := newTestSandboxAPIServer(t)
	defer srv.Close()
	c := newTestSandboxAPIClient(t, srv)

	err := c.DeleteSandbox(t.Context(), "sb_123")
	if err != nil {
		t.Fatalf("DeleteSandbox failed: %v", err)
	}
}

func TestSnapshotSandbox(t *testing.T) {
	srv := newTestSandboxAPIServer(t)
	defer srv.Close()
	c := newTestSandboxAPIClient(t, srv)

	resp, err := c.SnapshotSandbox(t.Context(), "sb_123", nil)
	if err != nil {
		t.Fatalf("SnapshotSandbox failed: %v", err)
	}
	if resp.SnapshotId != "snap_789" {
		t.Errorf("SnapshotId = %q, want %q", resp.SnapshotId, "snap_789")
	}
}

func TestSnapshotSandboxWithMode(t *testing.T) {
	srv := newTestSandboxAPIServer(t)
	defer srv.Close()
	c := newTestSandboxAPIClient(t, srv)

	resp, err := c.SnapshotSandbox(t.Context(), "sb_123", &SnapshotSandboxRequest{
		SnapshotContentMode: SnapshotContentModeFilesystemOnly,
	})
	if err != nil {
		t.Fatalf("SnapshotSandbox failed: %v", err)
	}
	if resp.SnapshotId != "snap_789" {
		t.Errorf("SnapshotId = %q, want %q", resp.SnapshotId, "snap_789")
	}
}

func TestSuspendSandbox(t *testing.T) {
	srv := newTestSandboxAPIServer(t)
	defer srv.Close()
	c := newTestSandboxAPIClient(t, srv)

	err := c.SuspendSandbox(t.Context(), "sb_123")
	if err != nil {
		t.Fatalf("SuspendSandbox failed: %v", err)
	}
}

func TestResumeSandbox(t *testing.T) {
	srv := newTestSandboxAPIServer(t)
	defer srv.Close()
	c := newTestSandboxAPIClient(t, srv)

	err := c.ResumeSandbox(t.Context(), "sb_123")
	if err != nil {
		t.Fatalf("ResumeSandbox failed: %v", err)
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
