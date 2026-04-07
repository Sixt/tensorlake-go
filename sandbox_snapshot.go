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

// SnapshotSandboxRequest holds options for snapshotting a sandbox.
type SnapshotSandboxRequest struct {
	SnapshotContentMode SnapshotContentMode `json:"snapshot_content_mode,omitempty"`
}

// SnapshotSandboxResponse represents the response from snapshotting a sandbox.
type SnapshotSandboxResponse struct {
	SnapshotId string `json:"snapshot_id"`
	Status     string `json:"status"`
}

// SnapshotSandbox creates a snapshot of a sandbox.
//
// See also: [Snapshot Sandbox API Reference]
//
// [Snapshot Sandbox API Reference]: https://docs.tensorlake.ai/api-reference/v2/sandboxes/snapshot
func (c *Client) SnapshotSandbox(ctx context.Context, sandboxID string, in *SnapshotSandboxRequest) (*SnapshotSandboxResponse, error) {
	var bodyReader io.Reader
	if in != nil {
		body, err := json.Marshal(in)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.sandboxAPIURL("/"+sandboxID+"/snapshot"), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return doSandboxAPI(c, req, func(r io.Reader) (*SnapshotSandboxResponse, error) {
		var result SnapshotSandboxResponse
		if err := json.NewDecoder(r).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return &result, nil
	})
}
