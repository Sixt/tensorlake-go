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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// GetDatasetRequest holds options for retrieving a dataset.
type GetDatasetRequest struct {
	// DatasetId is the unique identifier for the dataset.
	//
	// Required.
	DatasetId string

	// IncludeAnalytics includes parse job analytics in the response when set to true.
	//
	// Optional.
	IncludeAnalytics bool
}

// GetDataset retrieves details for a specific dataset.
//
// See also: [Get Dataset API Reference]
//
// [Get Dataset API Reference]: https://docs.tensorlake.ai/api-reference/v2/datasets/get
func (c *Client) GetDataset(ctx context.Context, in *GetDatasetRequest) (*Dataset, error) {
	reqURL := fmt.Sprintf("%s/datasets/%s", c.baseURL, in.DatasetId)
	if in.IncludeAnalytics {
		params := url.Values{}
		params.Add("include_analytics", "true")
		reqURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
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
