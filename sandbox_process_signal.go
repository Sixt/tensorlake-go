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

// SignalProcessRequest holds the signal to send to a process.
type SignalProcessRequest struct {
	Signal int32 `json:"signal"`
}

// SignalProcess sends a POSIX signal to a process in a sandbox.
//
// See also: [Signal Process API Reference]
//
// [Signal Process API Reference]: https://docs.tensorlake.ai/api-reference/v2/processes/signal
func (c *Client) SignalProcess(ctx context.Context, sandboxID string, pid int32, in *SignalProcessRequest) error {
	body, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	reqURL := fmt.Sprintf("%s/processes/%d/signal", c.sandboxProxyURL(sandboxID), pid)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := doSandbox(c, req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
