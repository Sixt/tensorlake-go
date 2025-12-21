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

// ParseDatasetRequest holds options for parsing a document with a dataset.
type ParseDatasetRequest struct {
	DatasetId string `json:"-"`
	FileSource
	PageRange string            `json:"page_range,omitempty"`
	FileName  string            `json:"file_name,omitempty"`
	MimeType  MimeType          `json:"mime_type,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// ParseDataset parses a document using a dataset's configuration.
func (c *Client) ParseDataset(ctx context.Context, in *ParseDatasetRequest) (*ParseJob, error) {
	if !in.SourceProvided() {
		return nil, fmt.Errorf("exactly one of file_id, file_url, or raw_text must be provided")
	}

	r, _ := json.Marshal(in)
	reqURL := fmt.Sprintf("%s/datasets/%s/parse", c.baseURL, in.DatasetId)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(r))
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
