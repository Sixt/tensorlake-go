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
)

// GetSandbox retrieves details for a specific sandbox.
//
// See also: [Get Sandbox API Reference]
//
// [Get Sandbox API Reference]: https://docs.tensorlake.ai/api-reference/v2/sandboxes/get
func (c *Client) GetSandbox(ctx context.Context, sandboxID string) (*SandboxInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.sandboxAPIURL("/"+sandboxID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return doSandboxAPI(c, req, func(r io.Reader) (*SandboxInfo, error) {
		var result SandboxInfo
		if err := json.NewDecoder(r).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return &result, nil
	})
}
