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
	"net/http"
)

// CloseProcessStdin closes a process's stdin stream (sends EOF).
//
// The process must have been started with StdinMode "pipe".
//
// See also: [Close Process Stdin API Reference]
//
// [Close Process Stdin API Reference]: https://docs.tensorlake.ai/api-reference/v2/processes/close-stdin
func (c *Client) CloseProcessStdin(ctx context.Context, sandboxID string, pid int32) error {
	reqURL := fmt.Sprintf("%s/processes/%d/stdin/close", c.sandboxProxyURL(sandboxID), pid)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := doSandbox(c, req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
