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
)

// GetParseResult retrieves the result of a parse job.
// The response will include: 1) parsed content (markdown or pages);
// 2) structured extraction results (if schemas are provided during the parse request);
// 3) page classification results (if page classifications are provided during the parse request).
//
// When the job finishes successfully, the response will contain pages
// (chunks of the page) chunks (text chunks extracted from the document),
// structured data (every schema_name provided in the parse request as a key).
func (c *Client) GetParseResult(ctx context.Context, parseId string) (*ParseResult, error) {
	reqURL := fmt.Sprintf("%s/parse/%s", c.baseURL, parseId)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return do(c, req, func(r io.Reader) (*ParseResult, error) {
		var result ParseResult
		if err := json.NewDecoder(r).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return &result, nil
	})
}
