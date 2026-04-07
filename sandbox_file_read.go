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

// ReadSandboxFile reads a file from a sandbox.
//
// The response is the raw file content as bytes.
//
// See also: [Read Sandbox File API Reference]
//
// [Read Sandbox File API Reference]: https://docs.tensorlake.ai/api-reference/v2/sandbox-files/read
func (c *Client) ReadSandboxFile(ctx context.Context, sandboxID, path string) ([]byte, error) {
	return readSandboxFileWithURL(c, ctx, c.sandboxProxyURL(sandboxID), path)
}

func readSandboxFileWithURL(c *Client, ctx context.Context, baseURL, path string) ([]byte, error) {
	reqURL := fmt.Sprintf("%s/files?path=%s", baseURL, url.QueryEscape(path))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := doSandbox(c, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return data, nil
}
