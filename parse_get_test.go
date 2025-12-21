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

func TestGetParseResultSSE(t *testing.T) {
	c := initializeTestClient(t)

	tests := []struct {
		req *ParseDocumentRequest
	}{
		{
			req: &ParseDocumentRequest{
				FileSource: FileSource{
					FileURL: "https://www.sixt.de/shared/t-c/sixt_DE_de.pdf",
				},
				ParsingOptions: &ParsingOptions{
					ChunkingStrategy: ChunkingStrategyNone,
				},
				EnrichmentOptions: &EnrichmentOptions{
					TableSummarization: true,
				},
				Labels: map[string]string{"category": "terms-and-conditions"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.req.FileSource.FileURL, func(t *testing.T) {
			func() {
				// Trigger read document operation.
				r, err := c.ParseDocument(t.Context(), tt.req)
				if err != nil {
					t.Fatalf("failed to parse document: %v", err)
				}
				if r == nil {
					t.Fatal("response is nil")
				}
				t.Logf("read document done, parse ID: %s", r.ParseId)

				// Read document status.
				result, err := c.GetParseResult(t.Context(), r.ParseId, WithSSE(true), WithOnUpdate(func(name ParseEventName, _ *ParseResult) {
					t.Logf("parse status: %s", name)
				}))
				if err != nil {
					t.Fatalf("failed to get parse result: %v", err)
				}
				if len(result.Chunks) == 0 {
					t.Fatalf("no chunks found")
				}
				peak := result.Chunks[0].Content
				if len(peak) > 100 {
					peak = peak[:100]
				}
				t.Logf("parse result: %+v", peak)

				// Validate parse results.
				jobs := []string{}
				for j, err := range c.IterParseJobs(t.Context(), 1) {
					if err != nil {
						t.Fatalf("failed to list parse jobs: %v", err)
					}
					jobs = append(jobs, j.ParseId)
				}
				t.Logf("listed %d parse jobs: %v", len(jobs), jobs)
				if !slices.Contains(jobs, r.ParseId) {
					t.Fatalf("parse job is not found in list: %v", jobs)
				}
				testCleanupFileAndParseJob(t, c, "", r.ParseId)
			}()
		})
	}
}
