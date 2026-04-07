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
	"testing"
	"time"
)

func TestSandboxProxyErrorFormat(t *testing.T) {
	tests := []struct {
		err  SandboxProxyError
		want string
	}{
		{
			err:  SandboxProxyError{Err: "not found"},
			want: "sandbox error: not found",
		},
		{
			err:  SandboxProxyError{Err: "forbidden", Code: "PATH_TRAVERSAL"},
			want: "sandbox error: forbidden (code: PATH_TRAVERSAL)",
		},
	}
	for _, tt := range tests {
		if got := tt.err.Error(); got != tt.want {
			t.Errorf("Error() = %q, want %q", got, tt.want)
		}
	}
}

func TestSandboxDirectoryListResponseDeserialization(t *testing.T) {
	raw := `{
		"path": "/workspace",
		"entries": [
			{"name": "src", "is_dir": true, "size": null, "modified_at": 1704067200000},
			{"name": "main.go", "is_dir": false, "size": 1234, "modified_at": 1704067200000}
		]
	}`

	var resp SandboxDirectoryListResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.Path != "/workspace" {
		t.Errorf("Path = %q, want %q", resp.Path, "/workspace")
	}
	if len(resp.Entries) != 2 {
		t.Fatalf("Entries length = %d, want 2", len(resp.Entries))
	}

	dir := resp.Entries[0]
	if dir.Name != "src" || !dir.IsDir {
		t.Errorf("first entry = %+v, want dir named src", dir)
	}
	if dir.Size != nil {
		t.Errorf("directory Size = %v, want nil", dir.Size)
	}

	file := resp.Entries[1]
	if file.Name != "main.go" || file.IsDir {
		t.Errorf("second entry = %+v, want file named main.go", file)
	}
	if file.Size == nil || *file.Size != 1234 {
		t.Errorf("file Size = %v, want 1234", file.Size)
	}
	if file.ModifiedAt == nil || *file.ModifiedAt != 1704067200000 {
		t.Errorf("file ModifiedAt = %v, want 1704067200000", file.ModifiedAt)
	}
}

func TestSandboxFileOperations(t *testing.T) {
	c := initializeSandboxClient(t)

	// Create a sandbox to operate on.
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

	// Wait for sandbox to be running.
	for range 30 {
		info, err := c.GetSandbox(t.Context(), sandboxID)
		if err != nil {
			t.Fatalf("failed to get sandbox: %v", err)
		}
		if info.Status == SandboxStatusRunning {
			break
		}
		t.Logf("sandbox status: %s, waiting...", info.Status)
		time.Sleep(2 * time.Second)
	}

	// Write a file.
	content := []byte("hello from integration test")
	err = c.WriteSandboxFile(t.Context(), sandboxID, "/workspace/test.txt", bytes.NewReader(content))
	if err != nil {
		t.Fatalf("failed to write sandbox file: %v", err)
	}
	t.Log("file written")

	// Read the file back.
	data, err := c.ReadSandboxFile(t.Context(), sandboxID, "/workspace/test.txt")
	if err != nil {
		t.Fatalf("failed to read sandbox file: %v", err)
	}
	if string(data) != string(content) {
		t.Errorf("file content = %q, want %q", string(data), string(content))
	}
	t.Log("file read back successfully")

	// List the directory.
	listResp, err := c.ListSandboxDirectory(t.Context(), sandboxID, "/workspace")
	if err != nil {
		t.Fatalf("failed to list sandbox directory: %v", err)
	}
	t.Logf("directory listing: %+v", listResp)

	found := false
	for _, entry := range listResp.Entries {
		if entry.Name == "test.txt" && !entry.IsDir {
			found = true
			break
		}
	}
	if !found {
		t.Error("test.txt not found in directory listing")
	}

	// Delete the file.
	err = c.DeleteSandboxFile(t.Context(), sandboxID, "/workspace/test.txt")
	if err != nil {
		t.Fatalf("failed to delete sandbox file: %v", err)
	}
	t.Log("file deleted")

	// Verify file is gone.
	_, err = c.ReadSandboxFile(t.Context(), sandboxID, "/workspace/test.txt")
	if err == nil {
		t.Error("expected error reading deleted file, got nil")
	}
}
