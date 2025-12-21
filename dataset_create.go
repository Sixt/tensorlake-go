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

// CreateDatasetRequest holds options for creating a dataset.
type CreateDatasetRequest struct {
	// The name of the dataset.
	//
	// The name can only contain alphanumeric characters, hyphens, and
	// underscores.
	//
	// The name must be unique within the organization and project context.
	//
	// Example:
	// "invoices dataset"
	Name string `json:"name"`

	// 	A description of the dataset.
	//
	// This field is optional and can be used to provide additional context
	// about the dataset.
	//
	// Example:
	// "This dataset contains all invoices from 2023."
	Description string `json:"description,omitempty"`

	// The properties of this object define the configuration for the document
	// parsing process.
	//
	// Tensorlake provides sane defaults that work well for most
	// documents, so this object is not required. However, every document
	// is different, and you may want to customize the parsing process to
	// better suit your needs.
	ParsingOptions *ParsingOptions `json:"parsing_options,omitempty"`

	// The properties of this object define the configuration for structured
	// data extraction.
	//
	// If this object is present, the API will perform structured data
	// extraction on the document.
	StructuredExtractionOptions []StructuredExtractionOptions `json:"structured_extraction_options,omitempty"`

	// The properties of this object define the configuration for page
	// classify.
	//
	// If this object is present, the API will perform page classify on
	// the document.
	PageClassifications []PageClassConfig `json:"page_classifications,omitempty"`

	// The properties of this object help to extend the output of the document
	// parsing process with additional information.
	//
	// This includes summarization of tables and figures, which can help to
	// provide a more comprehensive understanding of the document.
	//
	// This object is not required, and the API will use default settings if it
	// is not present.
	EnrichmentOptions *EnrichmentOptions `json:"enrichment_options,omitempty"`
}

// CreateDatasetResponse represents the response from creating a dataset.
type CreateDatasetResponse struct {
	// Name is the name of the dataset.
	Name string `json:"name"`
	// DatasetId is the ID of the created dataset.
	DatasetId string `json:"dataset_id"`
	// CreatedAt is the creation date and time of the dataset.
	CreatedAt string `json:"created_at"`
}

// CreateDataset creates a new dataset.
//
// See also: [Create Dataset API Reference]
//
// [Create Dataset API Reference]: https://docs.tensorlake.ai/api-reference/v2/datasets/create
func (c *Client) CreateDataset(ctx context.Context, in *CreateDatasetRequest) (*CreateDatasetResponse, error) {
	b, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/datasets", bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	return do(c, req, func(r io.Reader) (*CreateDatasetResponse, error) {
		var result CreateDatasetResponse
		if err := json.NewDecoder(r).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return &result, nil
	})
}
