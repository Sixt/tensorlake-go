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
	"log/slog"
	"net/http"
)

type ParseDocumentRequest struct {
	FileSource

	// ParsingOptions contains the properties of this object define
	// the configuration for the document parsing process.
	//
	// Tensorlake provides sane defaults that work well for most
	// documents, so this object is not required. However, every document
	// is different, and you may want to customize the parsing process to
	// better suit your needs.
	ParsingOptions *ParsingOptions `json:"parsing_options,omitempty"`

	// The properties of this object help to extend the output of the document
	// parsing process with additional information.
	//
	// This includes summarization of tables and figures, which can help to
	// provide a more comprehensive understanding of the document.
	//
	// This object is not required, and the API will use default settings if it
	// is not present.
	EnrichmentOptions *EnrichmentOptions `json:"enrichment_options,omitempty"`

	// StructuredExtractionOptions is the options for structured data extraction.
	//
	// The properties of this object define the configuration for structured
	// data extraction.
	//
	// If this object is present, the API will perform structured data
	// extraction on the document.
	StructuredExtractionOptions []StructuredExtractionOptions `json:"structured_extraction_options,omitempty"`

	// PageClassificationOptions is the options for page classification.
	//
	// The properties of this object define the configuration for page
	// classify.
	//
	// If this object is present, the API will perform page classify on
	// the document.
	PageClassificationOptions []PageClassConfig `json:"page_classifications,omitempty"`

	// PageRange is a comma-separated list of page numbers or
	// ranges to parse (e.g., '1,2,3-5'). Default: all pages.
	// Examples: "1-5,8,10"
	PageRange string `json:"page_range,omitempty"`

	// Additional metadata to identify the read request. The labels are
	// returned in the read response.
	Labels map[string]string `json:"labels,omitempty"`

	// MimeType is the MIME type of the file. This is used to determine how to process the file.
	MimeType MimeType `json:"mime_type,omitempty"`
}

// ParseDocument submits a document for comprehensive parsing (read, extract, and classify).
func (c *Client) ParseDocument(ctx context.Context, in *ParseDocumentRequest) (*ParseJob, error) {
	if !in.SourceProvided() {
		return nil, fmt.Errorf("exactly one of file_id, file_url, or raw_text must be provided")
	}

	body, _ := json.Marshal(in) // Impossible to fail?

	slog.Info("ParseDocument request", "request", string(body))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/parse", bytes.NewReader(body))
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
