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
	"testing"
)

func TestUsageDeserialization(t *testing.T) {
	raw := `{
		"pages_parsed": 5,
		"signature_detected_pages": 1,
		"strikethrough_detected_pages": 0,
		"ocr_input_tokens_used": 100,
		"ocr_output_tokens_used": 200,
		"extraction_input_tokens_used": 300,
		"extraction_output_tokens_used": 400,
		"summarization_input_tokens_used": 50,
		"summarization_output_tokens_used": 60
	}`

	var u Usage
	if err := json.Unmarshal([]byte(raw), &u); err != nil {
		t.Fatalf("failed to unmarshal Usage: %v", err)
	}

	if u.PagesParsed != 5 {
		t.Errorf("PagesParsed = %d, want 5", u.PagesParsed)
	}
	if u.OCRInputTokensUsed != 100 {
		t.Errorf("OCRInputTokensUsed = %d, want 100", u.OCRInputTokensUsed)
	}
	if u.OCROutputTokensUsed != 200 {
		t.Errorf("OCROutputTokensUsed = %d, want 200", u.OCROutputTokensUsed)
	}
	if u.ExtractionInputTokensUsed != 300 {
		t.Errorf("ExtractionInputTokensUsed = %d, want 300", u.ExtractionInputTokensUsed)
	}
	if u.ExtractionOutputTokensUsed != 400 {
		t.Errorf("ExtractionOutputTokensUsed = %d, want 400", u.ExtractionOutputTokensUsed)
	}
	if u.SummarizationInputTokensUsed != 50 {
		t.Errorf("SummarizationInputTokensUsed = %d, want 50", u.SummarizationInputTokensUsed)
	}
	if u.SummarizationOutputTokensUsed != 60 {
		t.Errorf("SummarizationOutputTokensUsed = %d, want 60", u.SummarizationOutputTokensUsed)
	}
}

func TestDatasetAnalyticsDeserialization(t *testing.T) {
	raw := `{
		"name": "test",
		"dataset_id": "ds_123",
		"status": "idle",
		"created_at": "2025-01-01T00:00:00Z",
		"updated_at": "2025-01-01T00:00:00Z",
		"analytics": {
			"total_processing_parse_jobs": 1,
			"total_pending_parse_jobs": 2,
			"total_error_parse_jobs": 3,
			"total_successful_parse_jobs": 4,
			"total_jobs": 10
		}
	}`

	var d Dataset
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		t.Fatalf("failed to unmarshal Dataset: %v", err)
	}

	if d.Analytics == nil {
		t.Fatal("Analytics is nil, want non-nil")
	}
	if d.Analytics.TotalJobs != 10 {
		t.Errorf("TotalJobs = %d, want 10", d.Analytics.TotalJobs)
	}
	if d.Analytics.TotalProcessingParseJobs != 1 {
		t.Errorf("TotalProcessingParseJobs = %d, want 1", d.Analytics.TotalProcessingParseJobs)
	}
	if d.Analytics.TotalPendingParseJobs != 2 {
		t.Errorf("TotalPendingParseJobs = %d, want 2", d.Analytics.TotalPendingParseJobs)
	}
	if d.Analytics.TotalErrorParseJobs != 3 {
		t.Errorf("TotalErrorParseJobs = %d, want 3", d.Analytics.TotalErrorParseJobs)
	}
	if d.Analytics.TotalSuccessfulParseJobs != 4 {
		t.Errorf("TotalSuccessfulParseJobs = %d, want 4", d.Analytics.TotalSuccessfulParseJobs)
	}
}

func TestDatasetAnalyticsDeserializationNull(t *testing.T) {
	raw := `{
		"name": "test",
		"dataset_id": "ds_123",
		"status": "idle",
		"created_at": "2025-01-01T00:00:00Z",
		"updated_at": "2025-01-01T00:00:00Z",
		"analytics": null
	}`

	var d Dataset
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		t.Fatalf("failed to unmarshal Dataset: %v", err)
	}

	if d.Analytics != nil {
		t.Errorf("Analytics = %+v, want nil", d.Analytics)
	}
}

func TestMergedTableDeserialization(t *testing.T) {
	raw := `{
		"parse_id": "parse_123",
		"parsed_pages_count": 3,
		"status": "successful",
		"created_at": "2025-01-01T00:00:00Z",
		"pages": [],
		"usage": {"pages_parsed": 3},
		"merged_tables": [
			{
				"merged_table_id": "mt_1",
				"merged_table_html": "<table><tr><td>A</td></tr></table>",
				"start_page": 1,
				"end_page": 3,
				"pages_merged": 3,
				"summary": "Revenue table",
				"merge_actions": {
					"pages": [1, 2, 3],
					"target_columns": 5
				}
			}
		]
	}`

	var r ParseResult
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		t.Fatalf("failed to unmarshal ParseResult: %v", err)
	}

	if len(r.MergedTables) != 1 {
		t.Fatalf("MergedTables length = %d, want 1", len(r.MergedTables))
	}

	mt := r.MergedTables[0]
	if mt.MergedTableId != "mt_1" {
		t.Errorf("MergedTableId = %q, want %q", mt.MergedTableId, "mt_1")
	}
	if mt.StartPage != 1 {
		t.Errorf("StartPage = %d, want 1", mt.StartPage)
	}
	if mt.EndPage != 3 {
		t.Errorf("EndPage = %d, want 3", mt.EndPage)
	}
	if mt.PagesMerged != 3 {
		t.Errorf("PagesMerged = %d, want 3", mt.PagesMerged)
	}
	if mt.Summary != "Revenue table" {
		t.Errorf("Summary = %q, want %q", mt.Summary, "Revenue table")
	}
	if mt.MergeActions == nil {
		t.Fatal("MergeActions is nil")
	}
	if len(mt.MergeActions.Pages) != 3 {
		t.Errorf("MergeActions.Pages length = %d, want 3", len(mt.MergeActions.Pages))
	}
	if mt.MergeActions.TargetColumns == nil || *mt.MergeActions.TargetColumns != 5 {
		t.Errorf("MergeActions.TargetColumns = %v, want 5", mt.MergeActions.TargetColumns)
	}
}

func TestPageClassificationReasonDeserialization(t *testing.T) {
	raw := `{
		"page_number": 1,
		"page_fragments": [],
		"classification_reason": "Contains invoice header"
	}`

	var p Page
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatalf("failed to unmarshal Page: %v", err)
	}

	if p.ClassificationReason != "Contains invoice header" {
		t.Errorf("ClassificationReason = %q, want %q", p.ClassificationReason, "Contains invoice header")
	}
}

func TestParseConfigurationDeserialization(t *testing.T) {
	raw := `{
		"parsing_options": {"chunking_strategy": "page", "merge_tables": true},
		"structured_extraction_options": [{"schema_name": "test", "json_schema": {}}],
		"page_classifications": [{"name": "invoice"}],
		"enrichment_options": {"table_cell_grounding": true, "chart_extraction": true, "key_value_extraction": true}
	}`

	var pc ParseConfiguration
	if err := json.Unmarshal([]byte(raw), &pc); err != nil {
		t.Fatalf("failed to unmarshal ParseConfiguration: %v", err)
	}

	if pc.ParsingOptions == nil {
		t.Fatal("ParsingOptions is nil")
	}
	if pc.ParsingOptions.ChunkingStrategy != ChunkingStrategyPage {
		t.Errorf("ChunkingStrategy = %q, want %q", pc.ParsingOptions.ChunkingStrategy, ChunkingStrategyPage)
	}
	if !pc.ParsingOptions.MergeTables {
		t.Error("MergeTables = false, want true")
	}
	if len(pc.StructuredExtractionOptions) != 1 {
		t.Errorf("StructuredExtractionOptions length = %d, want 1", len(pc.StructuredExtractionOptions))
	}
	if len(pc.PageClassifications) != 1 {
		t.Errorf("PageClassifications length = %d, want 1", len(pc.PageClassifications))
	}
	if pc.EnrichmentOptions == nil {
		t.Fatal("EnrichmentOptions is nil")
	}
	if !pc.EnrichmentOptions.TableCellGrounding {
		t.Error("TableCellGrounding = false, want true")
	}
	if !pc.EnrichmentOptions.ChartExtraction {
		t.Error("ChartExtraction = false, want true")
	}
	if !pc.EnrichmentOptions.KeyValueExtraction {
		t.Error("KeyValueExtraction = false, want true")
	}
}

func TestErrorResponseTimestampDeserialization(t *testing.T) {
	raw := `{
		"message": "not found",
		"code": "ENTITY_NOT_FOUND",
		"timestamp": 1704067200000,
		"trace_id": "abc123"
	}`

	var e ErrorResponse
	if err := json.Unmarshal([]byte(raw), &e); err != nil {
		t.Fatalf("failed to unmarshal ErrorResponse: %v", err)
	}

	if e.Timestamp != 1704067200000 {
		t.Errorf("Timestamp = %d, want 1704067200000", e.Timestamp)
	}
	if e.Code != ErrorCodeEntityNotFound {
		t.Errorf("Code = %q, want %q", e.Code, ErrorCodeEntityNotFound)
	}
}
