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
		files := fetchAllFiles(t, c)
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
		jobs := fetchAllParseJobs(t, c)
		if slices.Contains(jobs, parseId) {
			t.Fatalf("parse job is not deleted and found in list: %v", jobs)
		}
	}
}

func fetchAllFiles(t *testing.T, c *Client) []string {
	files, cursor := []string{}, ""
	for {
		listResp, err := c.ListFiles(t.Context(), &ListFilesRequest{
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
	return files
}

func fetchAllParseJobs(t *testing.T, c *Client) []string {
	jobs, cursor := []string{}, ""
	for {
		listResp, err := c.ListParseJobs(t.Context(), &ListParseJobsRequest{
			Cursor:    cursor,
			Limit:     1,
			Direction: PaginationDirectionNext,
		})
		if err != nil {
			t.Fatalf("failed to list parse jobs: %v", err)
		}
		if len(listResp.Items) == 0 {
			break
		}
		jobs = append(jobs, listResp.Items[0].ParseId)
		cursor = listResp.NextCursor

		if !listResp.HasMore {
			break
		}
	}
	return jobs
}

func fetchAllDatasets(t *testing.T, c *Client) []string {
	datasets, cursor := []string{}, ""
	for {
		listResp, err := c.ListDatasets(t.Context(), &ListDatasetsRequest{
			Cursor:    cursor,
			Limit:     1,
			Direction: PaginationDirectionNext,
		})
		if err != nil {
			t.Fatalf("failed to list datasets: %v", err)
		}
		if len(listResp.Items) == 0 {
			break
		}
		datasets = append(datasets, listResp.Items[0].DatasetId)
		cursor = listResp.NextCursor

		if !listResp.HasMore {
			break
		}
	}
	return datasets
}
