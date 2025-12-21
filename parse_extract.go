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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ExtractDocumentRequest holds options for extracting structured data from a document.
type ExtractDocumentRequest struct {
	FileSource

	StructuredExtractionOptions []StructuredExtractionOptions `json:"structured_extraction_options"`
	PageRange                   string                        `json:"page_range,omitempty"`
	MimeType                    string                        `json:"mime_type,omitempty"`
	Labels                      map[string]string             `json:"labels,omitempty"`
}

// ExtractDocument submits a document for structured data extraction.
//
// See also: [Extract Document API Reference]
//
// [Extract Document API Reference]: https://docs.tensorlake.ai/api-reference/v2/parse/extract
func (c *Client) ExtractDocument(ctx context.Context, in *ExtractDocumentRequest) (*ParseJob, error) {
	// Validate that exactly one source is provided
	if !in.SourceProvided() {
		return nil, fmt.Errorf("exactly one of file_id, file_url, or raw_text must be provided")
	}
	if len(in.StructuredExtractionOptions) == 0 {
		return nil, fmt.Errorf("at least one structured_extraction_options must be provided")
	}

	body, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/extract", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return do(c, req, func(r io.Reader) (*ParseJob, error) {
		var result ParseJob
		if err := json.NewDecoder(r).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return &result, nil
	})
}
