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
	"slices"
	"testing"
)

func TestClassifyDocument(t *testing.T) {
	c := initializeTestClient(t)

	tests := []struct {
		filepath string
		req      *ClassifyDocumentRequest
	}{
		{
			filepath: "testdata/acord.pdf",
			req: &ClassifyDocumentRequest{
				FileSource: FileSource{
					FileId: "", // fill later.
				},
				PageClassifications: []PageClassConfig{
					{
						Name:        "form125",
						Description: "ACORD 125: Applicant Information Section — captures general insured information, business details, and contacts",
					},
					{
						Name:        "form140",
						Description: "ACORD 140: Property Section — includes details about property coverage, location, valuation, and limit",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.filepath, func(t *testing.T) {
			func() {
				// Upload file.
				file, err := os.Open(tt.filepath)
				if err != nil {
					t.Fatalf("failed to open file: %v", err)
				}
				defer file.Close()

				resp, err := c.UploadFile(t.Context(), &UploadFileRequest{
					FileBytes: file,
					FileName:  filepath.Base(tt.filepath),
				})
				if err != nil {
					t.Fatalf("failed to upload file: %v", err)
				}
				if resp == nil {
					t.Fatal("response is nil")
				}
				if resp.FileId == "" {
					t.Fatal("file ID is empty")
				}

				tt.req.FileSource.FileId = resp.FileId

				// Classify document.
				r, err := c.ClassifyDocument(t.Context(), tt.req)
				if err != nil {
					t.Fatalf("failed to classify document: %v", err)
				}
				if r == nil {
					t.Fatal("response is nil")
				}
				t.Logf("classify document done, parse ID: %s", r.ParseId)

				// Get parse result.
				result, err := c.GetParseResult(t.Context(), r.ParseId, WithSSE(true), WithOnUpdate(func(name ParseEventName, _ *ParseResult) {
					t.Logf("parse status: %s", name)
				}))
				if err != nil {
					t.Fatalf("failed to get parse result: %v", err)
				}
				if result == nil {
					t.Fatal("response is nil")
				}
				if len(result.PageClasses) == 0 {
					t.Fatalf("no page classes found")
				}
				t.Logf("page classes: %v", result.PageClasses)

				// Validate classify results.
				jobs := []string{}
				for j, err := range c.IterParseJobs(t.Context(), 1) {
					if err != nil {
						t.Fatalf("failed to list parse jobs: %v", err)
					}
					jobs = append(jobs, j.ParseId)
				}
				t.Logf("parse jobs: %v", jobs)
				if !slices.Contains(jobs, r.ParseId) {
					t.Fatalf("parse job is not found in list: %v", jobs)
				}

				testCleanupFileAndParseJob(t, c, resp.FileId, r.ParseId)
			}()
		})
	}
}
