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
	"net/url"
)

// WriteSandboxFile writes a file to a sandbox.
//
// Parent directories are created automatically if they do not exist.
// The content is written as raw bytes.
//
// See also: [Write Sandbox File API Reference]
//
// [Write Sandbox File API Reference]: https://docs.tensorlake.ai/api-reference/v2/sandbox-files/write
func (c *Client) WriteSandboxFile(ctx context.Context, sandboxID, path string, content io.Reader) error {
	return writeSandboxFileWithURL(c, ctx, c.sandboxProxyURL(sandboxID), path, content)
}

func writeSandboxFileWithURL(c *Client, ctx context.Context, baseURL, path string, content io.Reader) error {
	reqURL := fmt.Sprintf("%s/files?path=%s", baseURL, url.QueryEscape(path))

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, content)
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
