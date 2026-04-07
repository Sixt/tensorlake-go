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
	"net/http"
	"net/http/httptest"
	"testing"
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

func newTestSandboxServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// List directory (must be registered before /api/v1/files to avoid prefix match)
	mux.HandleFunc("/api/v1/files/list", func(w http.ResponseWriter, r *http.Request) {
		size := int64(11)
		modAt := int64(1704067200000)
		json.NewEncoder(w).Encode(SandboxDirectoryListResponse{
			Path: r.URL.Query().Get("path"),
			Entries: []SandboxDirectoryEntry{
				{Name: "subdir", IsDir: true, ModifiedAt: &modAt},
				{Name: "hello.txt", IsDir: false, Size: &size, ModifiedAt: &modAt},
			},
		})
	})

	// Read / Write / Delete file
	mux.HandleFunc("/api/v1/files", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Query().Get("path")
		switch r.Method {
		case http.MethodGet:
			if path == "/workspace/hello.txt" {
				w.Header().Set("Content-Type", "application/octet-stream")
				w.Write([]byte("hello world"))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(SandboxProxyError{Err: "file not found"})
		case http.MethodPut:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			if path == "/workspace/hello.txt" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(SandboxProxyError{Err: "file not found"})
		}
	})

	return httptest.NewServer(mux)
}

func TestReadSandboxFile(t *testing.T) {
	srv := newTestSandboxServer(t)
	defer srv.Close()

	c := &Client{httpClient: srv.Client(), apiKey: "test-key"}
	data, err := readSandboxFileWithURL(c, t.Context(), srv.URL+"/api/v1", "/workspace/hello.txt")
	if err != nil {
		t.Fatalf("ReadSandboxFile failed: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("got %q, want %q", string(data), "hello world")
	}
}

func TestReadSandboxFileNotFound(t *testing.T) {
	srv := newTestSandboxServer(t)
	defer srv.Close()

	c := &Client{httpClient: srv.Client(), apiKey: "test-key"}
	_, err := readSandboxFileWithURL(c, t.Context(), srv.URL+"/api/v1", "/workspace/missing.txt")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	sandboxErr, ok := err.(*SandboxProxyError)
	if !ok {
		t.Fatalf("expected *SandboxProxyError, got %T: %v", err, err)
	}
	if sandboxErr.Err != "file not found" {
		t.Errorf("error = %q, want %q", sandboxErr.Err, "file not found")
	}
}

func TestWriteSandboxFile(t *testing.T) {
	srv := newTestSandboxServer(t)
	defer srv.Close()

	c := &Client{httpClient: srv.Client(), apiKey: "test-key"}
	err := writeSandboxFileWithURL(c, t.Context(), srv.URL+"/api/v1", "/workspace/new.txt", bytes.NewReader([]byte("content")))
	if err != nil {
		t.Fatalf("WriteSandboxFile failed: %v", err)
	}
}

func TestDeleteSandboxFile(t *testing.T) {
	srv := newTestSandboxServer(t)
	defer srv.Close()

	c := &Client{httpClient: srv.Client(), apiKey: "test-key"}
	err := deleteSandboxFileWithURL(c, t.Context(), srv.URL+"/api/v1", "/workspace/hello.txt")
	if err != nil {
		t.Fatalf("DeleteSandboxFile failed: %v", err)
	}
}

func TestDeleteSandboxFileNotFound(t *testing.T) {
	srv := newTestSandboxServer(t)
	defer srv.Close()

	c := &Client{httpClient: srv.Client(), apiKey: "test-key"}
	err := deleteSandboxFileWithURL(c, t.Context(), srv.URL+"/api/v1", "/workspace/missing.txt")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListSandboxDirectory(t *testing.T) {
	srv := newTestSandboxServer(t)
	defer srv.Close()

	c := &Client{httpClient: srv.Client(), apiKey: "test-key"}
	resp, err := listSandboxDirectoryWithURL(c, t.Context(), srv.URL+"/api/v1", "/workspace")
	if err != nil {
		t.Fatalf("ListSandboxDirectory failed: %v", err)
	}
	if resp.Path != "/workspace" {
		t.Errorf("Path = %q, want %q", resp.Path, "/workspace")
	}
	if len(resp.Entries) != 2 {
		t.Fatalf("Entries length = %d, want 2", len(resp.Entries))
	}
	if !resp.Entries[0].IsDir {
		t.Error("first entry should be a directory")
	}
	if resp.Entries[1].IsDir {
		t.Error("second entry should be a file")
	}
}
