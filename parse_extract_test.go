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
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
)

func TestExtractDocument(t *testing.T) {
	c := initializeTestClient(t)

	tests := []struct {
		filepath string
		req      *ExtractDocumentRequest
	}{
		{
			filepath: "testdata/bank_statement.pdf",
			req: &ExtractDocumentRequest{
				FileSource: FileSource{
					FileId: "", // fill later.
				},
				StructuredExtractionOptions: []StructuredExtractionOptions{
					{
						SchemaName: "form125-basic",
						JSONSchema: func() *jsonschema.Schema {
							s, err := jsonschema.For[BankStatement](nil)
							if err != nil {
								t.Fatalf("failed to get JSON schema: %v", err)
							}
							return s
						}(),
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

				// Upload file.
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

				// Parse document.
				r, err := c.ExtractDocument(t.Context(), tt.req)
				if err != nil {
					t.Fatalf("failed to extract document: %v", err)
				}
				if r == nil {
					t.Fatal("response is nil")
				}
				t.Logf("extract document done, parse ID: %s", r.ParseId)

				// Get parse result.
				result, err := c.GetParseResult(t.Context(), r.ParseId, WithSSE(true), WithOnUpdate(func(eventName string, _ *ParseResult) {
					t.Logf("parse status: %s", eventName)
				}))
				if err != nil {
					t.Fatalf("failed to get parse result: %v", err)
				}
				if result == nil {
					t.Fatal("response is nil")
				}
				if len(result.StructuredData) == 0 {
					t.Fatalf("no structured data found")
				}

				// Parse structured data using the provided JSON schema.
				data := result.StructuredData[0].Data
				var bank BankStatement
				if err := json.Unmarshal(data, &bank); err != nil {
					t.Fatalf("failed to unmarshal structured data: %v", err)
				}
				t.Logf("schema name: %s", result.StructuredData[0].SchemaName)
				t.Logf("page numbers: %v", result.StructuredData[0].PageNumbers)
				t.Logf("structured data: %+v", bank)

				// Validate parse results.
				jobs := []string{}
				for j, err := range c.IterParseJobs(t.Context(), 1, PaginationDirectionNext) {
					if err != nil {
						t.Fatalf("failed to list parse jobs: %v", err)
					}
					jobs = append(jobs, j.ParseId)
				}
				t.Logf("listed %d parse jobs: %v", len(jobs), jobs)
				if !slices.Contains(jobs, r.ParseId) {
					t.Fatalf("parse job is not found in list: %v", jobs)
				}

				testCleanupFileAndParseJob(t, c, resp.FileId, r.ParseId)
			}()
		})
	}
}
