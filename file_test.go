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
	"os"
	"path/filepath"
	"testing"
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
			files, err := fetchAllFiles(t, c)
			if err != nil {
				t.Fatalf("failed to list files: %v", err)
			}
			t.Logf("listed %d files: %v", len(files), files)

			found := false
			for _, file := range files {
				if file == resp.FileId {
					found = true
					break
				}
			}
			if !found {
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
			files, err = fetchAllFiles(t, c)
			if err != nil {
				t.Fatalf("failed to list files: %v", err)
			}
			found = false
			for _, file := range files {
				if file == resp.FileId {
					found = true
					break
				}
			}
			if found {
				t.Fatalf("file %s is not deleted", resp.FileId)
			}
		}()
	}
}

func fetchAllFiles(t *testing.T, c *Client) ([]string, error) {
	files, cursor := []string{}, ""
	var err error
	for {
		var listResp *PaginationResult[FileInfo]
		listResp, err = c.ListFiles(t.Context(), &ListFilesRequest{
			Cursor:    cursor,
			Limit:     1,
			Direction: PaginationDirectionNext,
		})
		if err != nil {
			t.Fatalf("failed to list files: %v", err)
		}
		if len(listResp.Items) == 0 {
			break
		}

		files = append(files, listResp.Items[0].FileId)
		cursor = listResp.NextCursor

		if !listResp.HasMore {
			break
		}
	}
	return files, err
}
