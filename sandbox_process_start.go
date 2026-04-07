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

// StartProcessRequest holds options for starting a process in a sandbox.
type StartProcessRequest struct {
	// Command is the executable to run.
	//
	// Required.
	Command string `json:"command"`

	// Args are command-line arguments.
	Args []string `json:"args,omitempty"`

	// Env sets environment variables for the process.
	Env map[string]string `json:"env,omitempty"`

	// WorkingDir is the working directory for the process.
	WorkingDir string `json:"working_dir,omitempty"`

	// StdinMode determines how stdin is handled. Default: "closed".
	StdinMode StdinMode `json:"stdin_mode,omitempty"`

	// StdoutMode determines how stdout is handled. Default: "capture".
	StdoutMode OutputMode `json:"stdout_mode,omitempty"`

	// StderrMode determines how stderr is handled. Default: "capture".
	StderrMode OutputMode `json:"stderr_mode,omitempty"`
}

// StartProcess starts a new process in a sandbox.
//
// See also: [Start Process API Reference]
//
// [Start Process API Reference]: https://docs.tensorlake.ai/api-reference/v2/processes/start
func (c *Client) StartProcess(ctx context.Context, sandboxID string, in *StartProcessRequest) (*ProcessInfo, error) {
	body, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	reqURL := c.sandboxProxyURL(sandboxID) + "/processes"
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

	var result ProcessInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}
