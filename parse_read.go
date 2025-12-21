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

// ReadDocumentRequest holds the input parameters for reading/parsing a document.
type ReadDocumentRequest struct {
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

	// Additional metadata to identify the read request. The labels are
	// returned in the read response.
	Labels map[string]string `json:"labels,omitempty"`

	// FileName is the name of the file. Only populated when using file_id.
	// Examples: "document.pdf"
	FileName string `json:"file_name,omitempty"`

	// PageRange is a comma-separated list of page numbers or
	// ranges to parse (e.g., '1,2,3-5'). Default: all pages.
	// Examples: "1-5,8,10"
	PageRange string `json:"page_range,omitempty"`

	// MimeType is the MIME type of the file. This is used to determine how to process the file.
	MimeType MimeType `json:"mime_type,omitempty"`
}

// ReadDocument submits an uploaded file, an internet-reachable URL, or
// any kind of raw text for document parsing. If you have configured a
// webhook, we will notify you when the job is complete, be it a success
// or a failure. The API will convert the document into markdown, and
// provide document layout information. Once submitted, the API will
// return a parse response with a parse_id field. You can query the status
// and results of the parse operation with the Get Parse Result endpoint.
func (c *Client) ReadDocument(ctx context.Context, in *ReadDocumentRequest) (*ParseJob, error) {
	if !in.SourceProvided() {
		return nil, fmt.Errorf("exactly one of file_id, file_url, or raw_text must be provided")
	}
	if in.FileId == "" {
		in.FileName = "" // FileName is only populated when using file_id.
	}

	body, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/read", bytes.NewReader(body))
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
