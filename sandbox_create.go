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

// CreateSandboxRequest holds options for creating a sandbox.
type CreateSandboxRequest struct {
	Name                       string                       `json:"name,omitempty"`
	Image                      string                       `json:"image,omitempty"`
	SnapshotId                 string                       `json:"snapshot_id,omitempty"`
	Entrypoint                 []string                     `json:"entrypoint,omitempty"`
	TimeoutSecs                *int64                       `json:"timeout_secs,omitempty"`
	SecretNames                []string                     `json:"secret_names,omitempty"`
	TemplateId                 string                       `json:"template_id,omitempty"`
	AllowUnauthenticatedAccess *bool                        `json:"allow_unauthenticated_access,omitempty"`
	ExposedPorts               []int32                      `json:"exposed_ports,omitempty"`
	Resources                  *SandboxResourceOverrides    `json:"resources,omitempty"`
	Network                    *SandboxNetworkAccessControl `json:"network,omitempty"`
}

// CreateSandboxResponse represents the response from creating a sandbox.
type CreateSandboxResponse struct {
	SandboxId     string               `json:"sandbox_id"`
	Status        SandboxStatus        `json:"status"`
	PendingReason SandboxPendingReason `json:"pending_reason,omitempty"`
}

// CreateSandbox creates a new sandbox.
//
// To restore from a snapshot, set SnapshotId in the request.
//
// See also: [Create Sandbox API Reference]
//
// [Create Sandbox API Reference]: https://docs.tensorlake.ai/api-reference/v2/sandboxes/create
func (c *Client) CreateSandbox(ctx context.Context, in *CreateSandboxRequest) (*CreateSandboxResponse, error) {
	body, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.sandboxAPIURL(""), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return doSandboxAPI(c, req, func(r io.Reader) (*CreateSandboxResponse, error) {
		var result CreateSandboxResponse
		if err := json.NewDecoder(r).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return &result, nil
	})
}
