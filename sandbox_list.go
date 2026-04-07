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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// ListSandboxesRequest holds options for listing sandboxes.
type ListSandboxesRequest struct {
	Limit     int    `json:"limit,omitempty"`
	Cursor    string `json:"cursor,omitempty"`
	Direction string `json:"direction,omitempty"`
	Status    string `json:"status,omitempty"`
}

// ListSandboxesResponse represents the response from listing sandboxes.
type ListSandboxesResponse struct {
	Sandboxes  []SandboxInfo `json:"sandboxes"`
	PrevCursor string        `json:"prev_cursor,omitempty"`
	NextCursor string        `json:"next_cursor,omitempty"`
}

// ListSandboxes lists sandboxes in the project.
//
// See also: [List Sandboxes API Reference]
//
// [List Sandboxes API Reference]: https://docs.tensorlake.ai/api-reference/v2/sandboxes/list
func (c *Client) ListSandboxes(ctx context.Context, in *ListSandboxesRequest) (*ListSandboxesResponse, error) {
	reqURL := c.sandboxAPIURL("")
	params := url.Values{}
	if in.Limit != 0 {
		params.Add("limit", fmt.Sprintf("%d", in.Limit))
	}
	if in.Cursor != "" {
		params.Add("cursor", in.Cursor)
	}
	if in.Direction != "" {
		params.Add("direction", in.Direction)
	}
	if in.Status != "" {
		params.Add("status", in.Status)
	}
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return doSandboxAPI(c, req, func(r io.Reader) (*ListSandboxesResponse, error) {
		var result ListSandboxesResponse
		if err := json.NewDecoder(r).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return &result, nil
	})
}
