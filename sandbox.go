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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	// DefaultSandboxProxyBaseURL is the default base URL for the sandbox file proxy.
	// The sandbox ID is prepended as a subdomain.
	DefaultSandboxProxyBaseURL = "https://sandbox.tensorlake.ai"
)

// SandboxProxyError represents an error returned by the sandbox proxy API.
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

// PTYSessionInfo represents metadata about a PTY session.
type PTYSessionInfo struct {
	SessionId string   `json:"session_id"`
	PID       int32    `json:"pid"`
	Command   string   `json:"command"`
	Args      []string `json:"args"`
	Rows      int32    `json:"rows"`
	Cols      int32    `json:"cols"`
	CreatedAt int64    `json:"created_at"`
	EndedAt   *int64   `json:"ended_at,omitempty"`
	ExitCode  *int32   `json:"exit_code,omitempty"`
	IsAlive   bool     `json:"is_alive"`
}

// sandboxProxyURL returns the base URL for sandbox file proxy requests.
// Format: https://{sandboxID}.{proxyHost}/api/v1
func (c *Client) sandboxProxyURL(sandboxID string) string {
	base := DefaultSandboxProxyBaseURL
	if c.sandboxProxyBaseURL != "" {
		base = c.sandboxProxyBaseURL
	}
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

// doSandbox executes a sandbox proxy API request and handles errors.
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
