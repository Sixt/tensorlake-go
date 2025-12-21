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

// UpdateDatasetRequest holds options for updating a dataset.
type UpdateDatasetRequest struct {
	DatasetId                   string                        `json:"-"`
	Description                 string                        `json:"description,omitempty"`
	ParsingOptions              *ParsingOptions               `json:"parsing_options,omitempty"`
	StructuredExtractionOptions []StructuredExtractionOptions `json:"structured_extraction_options,omitempty"`
	PageClassifications         []PageClassConfig             `json:"page_classifications,omitempty"`
	EnrichmentOptions           *EnrichmentOptions            `json:"enrichment_options,omitempty"`
}

// UpdateDataset updates a dataset's settings.
func (c *Client) UpdateDataset(ctx context.Context, in *UpdateDatasetRequest) (*Dataset, error) {
	reqBody, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	reqURL := fmt.Sprintf("%s/datasets/%s", c.baseURL, in.DatasetId)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return do(c, req, func(r io.Reader) (*Dataset, error) {
		var result Dataset
		if err := json.NewDecoder(r).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		return &result, nil
	})
}
