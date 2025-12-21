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
	"slices"
	"testing"
)

func testCleanupFileAndParseJob(t *testing.T, c *Client, fileId string, parseId string) {
	if fileId != "" {
		// Delete file.
		if err := c.DeleteFile(t.Context(), fileId); err != nil {
			t.Fatalf("failed to delete file: %v", err)
		}

		// Check if file is deleted.
		files := []string{}
		for f, err := range c.IterFiles(t.Context(), 1) {
			if err != nil {
				t.Fatalf("failed to list files: %v", err)
			}
			files = append(files, f.FileId)
		}
		t.Logf("listed %d files: %v", len(files), files)
		if slices.Contains(files, fileId) {
			t.Fatalf("file is not deleted and found in list: %v", files)
		}
	}

	if parseId != "" {
		// Delete parse job.
		if err := c.DeleteParseJob(t.Context(), parseId); err != nil {
			t.Fatalf("failed to delete parse job: %v", err)
		}

		// Check if parse job is deleted.
		jobs := []string{}
		for j, err := range c.IterParseJobs(t.Context(), 1) {
			if err != nil {
				t.Fatalf("failed to list parse jobs: %v", err)
			}
			jobs = append(jobs, j.ParseId)
		}
		t.Logf("listed %d parse jobs: %v", len(jobs), jobs)
		if slices.Contains(jobs, parseId) {
			t.Fatalf("parse job is not deleted and found in list: %v", jobs)
		}
	}
}
