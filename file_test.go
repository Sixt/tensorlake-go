// Copyright 2025 SIXT SE
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
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
	"time"
)

// TestFileManagement tests the file management functionality.
// It uploads a file, lists it, checks the metadata, and deletes it.
func TestFileManagement(t *testing.T) {
	c := initializeTestClient(t)

	tests := []struct {
		filepath string
		labels   map[string]string
	}{
		{
			filepath: "testdata/sixt_DE_de.pdf",
			labels:   map[string]string{"category": "terms-and-conditions"},
		},
	}
	for _, tt := range tests {
		func() {
			// Open the file.
			file, err := os.Open(tt.filepath)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()

			fileSize, err := file.Stat()
			if err != nil {
				t.Fatal(err)
			}

			// Upload the file.
			resp, err := c.UploadFile(t.Context(), &UploadFileRequest{
				FileBytes: file,
				FileName:  filepath.Base(tt.filepath),
				Labels:    tt.labels,
			})
			if err != nil {
				t.Fatalf("failed to upload file: %v", err)
			}

			// Validate the response.
			if resp == nil {
				t.Fatal("response is nil")
			}
			if resp.FileId == "" {
				t.Fatal("file ID is empty")
			}
			if resp.CreatedAt.IsZero() {
				t.Fatal("created at is zero")
			}

			t.Log("upload file done, begin listing files...")

			// List the files. Iterate through all the pages.
			files := []string{}
			for f, err := range c.IterFiles(t.Context(), 1) {
				if err != nil {
					t.Fatalf("failed to list files: %v", err)
				}
				files = append(files, f.FileId)
			}
			t.Logf("listed %d files: %v", len(files), files)

			if !slices.Contains(files, resp.FileId) {
				t.Fatalf("file %s not found in list", resp.FileId)
			}

			// Check file metadata.
			metaResp, err := c.GetFileMetadata(t.Context(), resp.FileId)
			if err != nil {
				t.Fatalf("failed to get file metadata: %v", err)
			}
			if metaResp == nil {
				t.Fatal("metadata response is nil")
			}
			if metaResp.FileId != resp.FileId {
				t.Fatalf("file ID mismatch: %s != %s", metaResp.FileId, resp.FileId)
			}
			if metaResp.FileName != filepath.Base(tt.filepath) {
				t.Fatalf("file name mismatch: %s != %s", metaResp.FileName, filepath.Base(tt.filepath))
			}
			if metaResp.MimeType != MimeTypePDF {
				t.Fatalf("mime type mismatch: %s != %s", metaResp.MimeType, MimeTypePDF)
			}
			if metaResp.FileSize != fileSize.Size() {
				t.Fatalf("file size mismatch: %d != %d", metaResp.FileSize, fileSize.Size())
			}
			if metaResp.CreatedAt == "" {
				t.Fatal("created at is zero")
			}
			t.Logf("file metadata: %+v", metaResp)

			// Delete the file.
			if err := c.DeleteFile(t.Context(), resp.FileId); err != nil {
				t.Fatalf("failed to delete file: %v", err)
			}

			// Validate file is deleted.
			files = []string{}
			for f, err := range c.IterFiles(t.Context(), 1) {
				if err != nil {
					t.Fatalf("failed to list files: %v", err)
				}
				files = append(files, f.FileId)
			}
			t.Logf("listed %d files: %v", len(files), files)
			if slices.Contains(files, resp.FileId) {
				t.Fatalf("file %s is not deleted", resp.FileId)
			}
		}()
	}
}

// TestUploadFileNoGoroutineLeak tests that the UploadFile function
// doesn't leak goroutines when the context is cancelled before the
// HTTP request can be executed.
func TestUploadFileNoGoroutineLeak(t *testing.T) {
	// Count goroutines before
	before := runtime.NumGoroutine()

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Create a client with a cancelled context scenario
	c := NewClient(
		WithBaseURL("http://localhost:9999"),
		WithAPIKey("test-key"),
	)

	// Try to upload a file with cancelled context
	_, err := c.UploadFile(ctx, &UploadFileRequest{
		FileBytes: bytes.NewReader([]byte("test data")),
		FileName:  "test.txt",
	})

	// Should get an error
	if err == nil {
		t.Fatal("expected error with cancelled context")
	}

	// Give any leaked goroutines a moment to manifest
	time.Sleep(100 * time.Millisecond)

	// Count goroutines after
	after := runtime.NumGoroutine()

	// Allow for some variance (the exact count can fluctuate)
	// but we shouldn't have leaked more than 1 goroutine
	if after > before+1 {
		t.Errorf("possible goroutine leak: before=%d, after=%d, leaked=%d", before, after, after-before)
	}
}

// TestUploadFileWithHTTPError tests that the UploadFile function
// properly cleans up when the HTTP request fails.
func TestUploadFileWithHTTPError(t *testing.T) {
	// Count goroutines before
	before := runtime.NumGoroutine()

	// Create a custom HTTP client that always returns an error
	errorClient := &http.Client{
		Transport: &errorRoundTripper{},
	}

	c := NewClient(
		WithBaseURL("http://localhost:9999"),
		WithAPIKey("test-key"),
		WithHTTPClient(errorClient),
	)

	// Try to upload a file
	_, err := c.UploadFile(context.Background(), &UploadFileRequest{
		FileBytes: bytes.NewReader([]byte("test data")),
		FileName:  "test.txt",
	})

	// Should get an error
	if err == nil {
		t.Fatal("expected error with failing HTTP client")
	}

	// Give any leaked goroutines a moment to manifest
	time.Sleep(100 * time.Millisecond)

	// Count goroutines after
	after := runtime.NumGoroutine()

	// Allow for some variance but we shouldn't have leaked goroutines
	if after > before+1 {
		t.Errorf("possible goroutine leak: before=%d, after=%d, leaked=%d", before, after, after-before)
	}
}

// errorRoundTripper is a http.RoundTripper that always returns an error
type errorRoundTripper struct{}

func (e *errorRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, errors.New("simulated network error")
}
