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
	"net/http"
)

// ProcessListResponse represents the response from listing processes.
type ProcessListResponse struct {
	Processes []ProcessInfo `json:"processes"`
}

// ListProcesses lists all tracked processes in a sandbox.
//
// See also: [List Processes API Reference]
//
// [List Processes API Reference]: https://docs.tensorlake.ai/api-reference/v2/processes/list
func (c *Client) ListProcesses(ctx context.Context, sandboxID string) (*ProcessListResponse, error) {
	reqURL := c.sandboxProxyURL(sandboxID) + "/processes"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := doSandbox(c, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ProcessListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}
