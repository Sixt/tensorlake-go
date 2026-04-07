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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// SandboxProxyError represents an error returned by the sandbox API.
type SandboxProxyError struct {
	// Err is a human-readable error message.
	Err string `json:"error"`
	// Code is an optional machine-readable error code.
	Code string `json:"code,omitempty"`
}

func (e *SandboxProxyError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("sandbox error: %s (code: %s)", e.Err, e.Code)
	}
	return fmt.Sprintf("sandbox error: %s", e.Err)
}

// SandboxDirectoryEntry represents a file or directory entry in a sandbox.
type SandboxDirectoryEntry struct {
	// Name is the name of the file or directory.
	Name string `json:"name"`
	// IsDir indicates whether this entry is a directory.
	IsDir bool `json:"is_dir"`
	// Size is the file size in bytes. Nil for directories.
	Size *int64 `json:"size,omitempty"`
	// ModifiedAt is the last modification time in milliseconds since epoch. Nil if unavailable.
	ModifiedAt *int64 `json:"modified_at,omitempty"`
}

// SandboxDirectoryListResponse represents the response from listing a sandbox directory.
type SandboxDirectoryListResponse struct {
	// Path is the directory path that was listed.
	Path string `json:"path"`
	// Entries contains the directory entries, sorted with directories first, then alphabetically.
	Entries []SandboxDirectoryEntry `json:"entries"`
}

const (
	// DefaultSandboxProxyBaseURL is the default base URL for the sandbox file proxy.
	// The sandbox ID is prepended as a subdomain.
	DefaultSandboxProxyBaseURL = "https://sandbox.tensorlake.ai"
)

// sandboxProxyURL returns the base URL for sandbox file proxy requests.
// Format: https://{sandboxID}.{proxyHost}/api/v1
func (c *Client) sandboxProxyURL(sandboxID string) string {
	base := DefaultSandboxProxyBaseURL
	if c.sandboxProxyBaseURL != "" {
		base = c.sandboxProxyBaseURL
	}
	// Insert sandboxID as subdomain: https://sandbox.example.com → https://{id}.sandbox.example.com
	return fmt.Sprintf("https://%s.%s/api/v1", sandboxID, stripScheme(base))
}

// stripScheme removes the "https://" or "http://" prefix from a URL.
func stripScheme(u string) string {
	for _, prefix := range []string{"https://", "http://"} {
		if len(u) > len(prefix) && u[:len(prefix)] == prefix {
			return u[len(prefix):]
		}
	}
	return u
}

// doSandbox executes a sandbox API request and handles errors.
func doSandbox(c *Client, req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		var sandboxErr SandboxProxyError
		if err := json.NewDecoder(resp.Body).Decode(&sandboxErr); err != nil {
			bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
			return nil, fmt.Errorf("sandbox request failed (%d): %s", resp.StatusCode, string(bodyBytes))
		}
		return nil, &sandboxErr
	}

	return resp, nil
}

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

// ListSandboxDirectory lists the contents of a directory in a sandbox.
//
// Entries are sorted with directories first, then alphabetically.
//
// See also: [List Sandbox Directory API Reference]
//
// [List Sandbox Directory API Reference]: https://docs.tensorlake.ai/api-reference/v2/sandbox-files/list
func (c *Client) ListSandboxDirectory(ctx context.Context, sandboxID, path string) (*SandboxDirectoryListResponse, error) {
	return listSandboxDirectoryWithURL(c, ctx, c.sandboxProxyURL(sandboxID), path)
}

func listSandboxDirectoryWithURL(c *Client, ctx context.Context, baseURL, path string) (*SandboxDirectoryListResponse, error) {
	reqURL := fmt.Sprintf("%s/files/list?path=%s", baseURL, url.QueryEscape(path))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := doSandbox(c, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result SandboxDirectoryListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}
