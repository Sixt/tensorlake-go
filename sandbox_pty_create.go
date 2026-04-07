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
	"net/http"
)

// CreatePTYRequest holds options for creating a PTY session.
type CreatePTYRequest struct {
	// Command is the executable to run (e.g. "/bin/bash").
	//
	// Required.
	Command string `json:"command"`

	// Args are command-line arguments (e.g. ["-l"]).
	Args []string `json:"args,omitempty"`

	// Env sets environment variables for the session.
	Env map[string]string `json:"env,omitempty"`

	// WorkingDir is the initial working directory.
	WorkingDir string `json:"working_dir,omitempty"`

	// Rows is the terminal height. Default: 24. Clamped to 1..500.
	Rows int32 `json:"rows,omitempty"`

	// Cols is the terminal width. Default: 80. Clamped to 1..1000.
	Cols int32 `json:"cols,omitempty"`
}

// CreatePTYResponse represents the response from creating a PTY session.
type CreatePTYResponse struct {
	// SessionId is the unique PTY session identifier.
	SessionId string `json:"session_id"`

	// Token is used for WebSocket connection authentication.
	Token string `json:"token"`
}

// CreatePTY creates a new PTY session in a sandbox.
//
// Returns a session ID and token for WebSocket authentication.
// The maximum number of concurrent PTY sessions per sandbox is 64.
//
// See also: [Create PTY Session API Reference]
//
// [Create PTY Session API Reference]: https://docs.tensorlake.ai/api-reference/v2/pty/create
func (c *Client) CreatePTY(ctx context.Context, sandboxID string, in *CreatePTYRequest) (*CreatePTYResponse, error) {
	body, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	reqURL := c.sandboxProxyURL(sandboxID) + "/pty"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := doSandbox(c, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result CreatePTYResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}
