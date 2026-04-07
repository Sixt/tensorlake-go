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
	"fmt"
	"io"
	"net/http"
)

// WriteProcessStdin writes raw bytes to a process's stdin.
//
// The process must have been started with StdinMode "pipe".
//
// See also: [Write Process Stdin API Reference]
//
// [Write Process Stdin API Reference]: https://docs.tensorlake.ai/api-reference/v2/processes/stdin
func (c *Client) WriteProcessStdin(ctx context.Context, sandboxID string, pid int32, data io.Reader) error {
	reqURL := fmt.Sprintf("%s/processes/%d/stdin", c.sandboxProxyURL(sandboxID), pid)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, data)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := doSandbox(c, req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
