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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// UpdateSandboxRequest holds options for updating a sandbox.
type UpdateSandboxRequest struct {
	AllowUnauthenticatedAccess *bool   `json:"allow_unauthenticated_access,omitempty"`
	ExposedPorts               []int32 `json:"exposed_ports,omitempty"`
}

// UpdateSandbox updates a sandbox's settings.
//
// See also: [Update Sandbox API Reference]
//
// [Update Sandbox API Reference]: https://docs.tensorlake.ai/api-reference/v2/sandboxes/update
func (c *Client) UpdateSandbox(ctx context.Context, sandboxID string, in *UpdateSandboxRequest) (*SandboxInfo, error) {
	body, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.sandboxAPIURL("/"+sandboxID), bytes.NewReader(body))
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
