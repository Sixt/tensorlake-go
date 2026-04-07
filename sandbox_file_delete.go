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
	"net/url"
)

// DeleteSandboxFile deletes a file from a sandbox.
//
// See also: [Delete Sandbox File API Reference]
//
// [Delete Sandbox File API Reference]: https://docs.tensorlake.ai/api-reference/v2/sandbox-files/delete
func (c *Client) DeleteSandboxFile(ctx context.Context, sandboxID, path string) error {
	return deleteSandboxFileWithURL(c, ctx, c.sandboxProxyURL(sandboxID), path)
}

func deleteSandboxFileWithURL(c *Client, ctx context.Context, baseURL, path string) error {
	reqURL := fmt.Sprintf("%s/files?path=%s", baseURL, url.QueryEscape(path))

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
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
