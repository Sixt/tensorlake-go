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

// ResizePTYRequest holds the terminal dimensions for a resize operation.
type ResizePTYRequest struct {
	Rows int32 `json:"rows"`
	Cols int32 `json:"cols"`
}

// ResizePTY resizes a PTY session's terminal dimensions.
//
// Rows are clamped to 1..500, cols to 1..1000 server-side.
//
// See also: [Resize PTY Session API Reference]
//
// [Resize PTY Session API Reference]: https://docs.tensorlake.ai/api-reference/v2/pty/resize
func (c *Client) ResizePTY(ctx context.Context, sandboxID, sessionID string, in *ResizePTYRequest) error {
	body, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	reqURL := c.sandboxProxyURL(sandboxID) + "/pty/" + sessionID + "/resize"
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
