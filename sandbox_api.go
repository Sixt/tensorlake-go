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
	"io"
	"net/http"
	"net/url"
)

const (
	// SandboxAPIBaseURL is the base URL for sandbox management operations.
	SandboxAPIBaseURL = "https://api.tensorlake.ai"
)

// SandboxStatus represents the current state of a sandbox.
type SandboxStatus string

const (
	SandboxStatusPending      SandboxStatus = "pending"
	SandboxStatusRunning      SandboxStatus = "running"
	SandboxStatusSnapshotting SandboxStatus = "snapshotting"
	SandboxStatusSuspending   SandboxStatus = "suspending"
	SandboxStatusSuspended    SandboxStatus = "suspended"
	SandboxStatusTerminated   SandboxStatus = "terminated"
)

// SandboxPendingReason describes why a sandbox is in pending state.
type SandboxPendingReason string

const (
	SandboxPendingReasonScheduling           SandboxPendingReason = "scheduling"
	SandboxPendingReasonWaitingForContainer  SandboxPendingReason = "waiting_for_container"
	SandboxPendingReasonNoExecutorsAvailable SandboxPendingReason = "no_executors_available"
	SandboxPendingReasonNoResourcesAvailable SandboxPendingReason = "no_resources_available"
	SandboxPendingReasonPoolAtCapacity       SandboxPendingReason = "pool_at_capacity"
)

// SnapshotContentMode determines what content is captured in a snapshot.
type SnapshotContentMode string

const (
	SnapshotContentModeFull           SnapshotContentMode = "full"
	SnapshotContentModeFilesystemOnly SnapshotContentMode = "filesystem_only"
)

// SandboxInfo represents detailed information about a sandbox.
type SandboxInfo struct {
	Id                          string                       `json:"id"`
	Namespace                   string                       `json:"namespace"`
	Image                       string                       `json:"image,omitempty"`
	Status                      SandboxStatus                `json:"status"`
	PendingReason               string                       `json:"pending_reason,omitempty"`
	Outcome                     string                       `json:"outcome,omitempty"`
	CreatedAt                   int64                        `json:"created_at"`
	ContainerId                 string                       `json:"container_id,omitempty"`
	ExecutorId                  string                       `json:"executor_id,omitempty"`
	Resources                   ContainerResourcesInfo       `json:"resources"`
	TimeoutSecs                 int64                        `json:"timeout_secs"`
	SandboxURL                  string                       `json:"sandbox_url,omitempty"`
	PoolId                      string                       `json:"pool_id,omitempty"`
	NetworkPolicy               *SandboxNetworkAccessControl `json:"network_policy,omitempty"`
	AllowUnauthenticatedAccess  bool                         `json:"allow_unauthenticated_access"`
	ExposedPorts                []int32                      `json:"exposed_ports,omitempty"`
	TemplateId                  string                       `json:"template_id,omitempty"`
	Name                        string                       `json:"name,omitempty"`
}

// ContainerResourcesInfo describes the resource allocation of a sandbox.
type ContainerResourcesInfo struct {
	CPUs            float64 `json:"cpus"`
	MemoryMB        int64   `json:"memory_mb"`
	EphemeralDiskMB int64   `json:"ephemeral_disk_mb"`
}

// SandboxNetworkAccessControl configures network access for a sandbox.
type SandboxNetworkAccessControl struct {
	AllowInternetAccess bool     `json:"allow_internet_access"`
	AllowOut            []string `json:"allow_out,omitempty"`
	DenyOut             []string `json:"deny_out,omitempty"`
}

// GPUResources specifies GPU allocation.
type GPUResources struct {
	Count int32  `json:"count"`
	Model string `json:"model"`
}

// SandboxResourceOverrides configures resource allocation for a sandbox.
type SandboxResourceOverrides struct {
	CPUs            float64        `json:"cpus,omitempty"`
	MemoryMB        int64          `json:"memory_mb,omitempty"`
	EphemeralDiskMB int64          `json:"ephemeral_disk_mb,omitempty"` // Deprecated: server ignores this field.
	GPUs            []GPUResources `json:"gpus,omitempty"`
}

// CreateSandboxRequest holds options for creating a sandbox.
type CreateSandboxRequest struct {
	Name                       string                       `json:"name,omitempty"`
	Image                      string                       `json:"image,omitempty"`
	SnapshotId                 string                       `json:"snapshot_id,omitempty"`
	Entrypoint                 []string                     `json:"entrypoint,omitempty"`
	TimeoutSecs                *int64                       `json:"timeout_secs,omitempty"`
	SecretNames                []string                     `json:"secret_names,omitempty"`
	TemplateId                 string                       `json:"template_id,omitempty"`
	AllowUnauthenticatedAccess *bool                        `json:"allow_unauthenticated_access,omitempty"`
	ExposedPorts               []int32                      `json:"exposed_ports,omitempty"`
	Resources                  *SandboxResourceOverrides    `json:"resources,omitempty"`
	Network                    *SandboxNetworkAccessControl `json:"network,omitempty"`
}

// CreateSandboxResponse represents the response from creating a sandbox.
type CreateSandboxResponse struct {
	SandboxId     string               `json:"sandbox_id"`
	Status        SandboxStatus        `json:"status"`
	PendingReason SandboxPendingReason `json:"pending_reason,omitempty"`
}

// UpdateSandboxRequest holds options for updating a sandbox.
type UpdateSandboxRequest struct {
	AllowUnauthenticatedAccess *bool   `json:"allow_unauthenticated_access,omitempty"`
	ExposedPorts               []int32 `json:"exposed_ports,omitempty"`
}

// ListSandboxesRequest holds options for listing sandboxes.
type ListSandboxesRequest struct {
	Limit     int    `json:"limit,omitempty"`
	Cursor    string `json:"cursor,omitempty"`
	Direction string `json:"direction,omitempty"`
	Status    string `json:"status,omitempty"`
}

// ListSandboxesResponse represents the response from listing sandboxes.
type ListSandboxesResponse struct {
	Sandboxes  []SandboxInfo `json:"sandboxes"`
	PrevCursor string        `json:"prev_cursor,omitempty"`
	NextCursor string        `json:"next_cursor,omitempty"`
}

// SnapshotSandboxRequest holds options for snapshotting a sandbox.
type SnapshotSandboxRequest struct {
	SnapshotContentMode SnapshotContentMode `json:"snapshot_content_mode,omitempty"`
}

// SnapshotSandboxResponse represents the response from snapshotting a sandbox.
type SnapshotSandboxResponse struct {
	SnapshotId string `json:"snapshot_id"`
	Status     string `json:"status"`
}

// sandboxAPIURL constructs a URL for sandbox management API calls.
func (c *Client) sandboxAPIURL(path string) string {
	base := SandboxAPIBaseURL
	if c.sandboxAPIBaseURL != "" {
		base = c.sandboxAPIBaseURL
	}
	return base + path
}

// doSandboxAPI executes a sandbox management API request.
// Unlike doSandbox (for sandbox-proxy file ops), errors here are text/plain.
func doSandboxAPI[T any](c *Client, req *http.Request, successHandler func(io.Reader) (T, error)) (T, error) {
	var zero T

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	if req.Body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return zero, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		if successHandler != nil {
			return successHandler(resp.Body)
		}
		return zero, nil

	default:
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return zero, fmt.Errorf("sandbox API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}
}

// CreateSandbox creates a new sandbox.
//
// To restore from a snapshot, set SnapshotId in the request.
//
// See also: [Create Sandbox API Reference]
//
// [Create Sandbox API Reference]: https://docs.tensorlake.ai/api-reference/v2/sandboxes/create
func (c *Client) CreateSandbox(ctx context.Context, in *CreateSandboxRequest) (*CreateSandboxResponse, error) {
	body, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.sandboxAPIURL("/sandboxes"), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return doSandboxAPI(c, req, func(r io.Reader) (*CreateSandboxResponse, error) {
		var result CreateSandboxResponse
		if err := json.NewDecoder(r).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return &result, nil
	})
}

// ListSandboxes lists sandboxes in the project.
//
// See also: [List Sandboxes API Reference]
//
// [List Sandboxes API Reference]: https://docs.tensorlake.ai/api-reference/v2/sandboxes/list
func (c *Client) ListSandboxes(ctx context.Context, in *ListSandboxesRequest) (*ListSandboxesResponse, error) {
	reqURL := c.sandboxAPIURL("/sandboxes")
	params := url.Values{}
	if in.Limit != 0 {
		params.Add("limit", fmt.Sprintf("%d", in.Limit))
	}
	if in.Cursor != "" {
		params.Add("cursor", in.Cursor)
	}
	if in.Direction != "" {
		params.Add("direction", in.Direction)
	}
	if in.Status != "" {
		params.Add("status", in.Status)
	}
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return doSandboxAPI(c, req, func(r io.Reader) (*ListSandboxesResponse, error) {
		var result ListSandboxesResponse
		if err := json.NewDecoder(r).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return &result, nil
	})
}

// GetSandbox retrieves details for a specific sandbox.
//
// See also: [Get Sandbox API Reference]
//
// [Get Sandbox API Reference]: https://docs.tensorlake.ai/api-reference/v2/sandboxes/get
func (c *Client) GetSandbox(ctx context.Context, sandboxID string) (*SandboxInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.sandboxAPIURL("/sandboxes/"+sandboxID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return doSandboxAPI(c, req, func(r io.Reader) (*SandboxInfo, error) {
		var result SandboxInfo
		if err := json.NewDecoder(r).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return &result, nil
	})
}

// UpdateSandbox updates a sandbox's settings.
//
// See also: [Update Sandbox API Reference]
//
// [Update Sandbox API Reference]: https://docs.tensorlake.ai/api-reference/v2/sandboxes/update
func (c *Client) UpdateSandbox(ctx context.Context, sandboxID string, in *UpdateSandboxRequest) (*SandboxInfo, error) {
	body, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.sandboxAPIURL("/sandboxes/"+sandboxID), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return doSandboxAPI(c, req, func(r io.Reader) (*SandboxInfo, error) {
		var result SandboxInfo
		if err := json.NewDecoder(r).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return &result, nil
	})
}

// DeleteSandbox terminates a sandbox.
//
// This operation is idempotent — terminating an already-terminated sandbox returns success.
//
// See also: [Delete Sandbox API Reference]
//
// [Delete Sandbox API Reference]: https://docs.tensorlake.ai/api-reference/v2/sandboxes/delete
func (c *Client) DeleteSandbox(ctx context.Context, sandboxID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.sandboxAPIURL("/sandboxes/"+sandboxID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = doSandboxAPI[struct{}](c, req, nil)
	return err
}

// SnapshotSandbox creates a snapshot of a sandbox.
//
// See also: [Snapshot Sandbox API Reference]
//
// [Snapshot Sandbox API Reference]: https://docs.tensorlake.ai/api-reference/v2/sandboxes/snapshot
func (c *Client) SnapshotSandbox(ctx context.Context, sandboxID string, in *SnapshotSandboxRequest) (*SnapshotSandboxResponse, error) {
	var bodyReader io.Reader
	if in != nil {
		body, err := json.Marshal(in)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.sandboxAPIURL("/sandboxes/"+sandboxID+"/snapshot"), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return doSandboxAPI(c, req, func(r io.Reader) (*SnapshotSandboxResponse, error) {
		var result SnapshotSandboxResponse
		if err := json.NewDecoder(r).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return &result, nil
	})
}

// SuspendSandbox suspends a named sandbox.
//
// Only named sandboxes can be suspended. Ephemeral sandboxes return an error.
// Returns nil on success (both 200 already-suspended and 202 suspend-initiated).
//
// See also: [Suspend Sandbox API Reference]
//
// [Suspend Sandbox API Reference]: https://docs.tensorlake.ai/api-reference/v2/sandboxes/suspend
func (c *Client) SuspendSandbox(ctx context.Context, sandboxID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.sandboxAPIURL("/sandboxes/"+sandboxID+"/suspend"), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = doSandboxAPI[struct{}](c, req, nil)
	return err
}

// ResumeSandbox resumes a suspended sandbox.
//
// Returns nil on success (both 200 already-running and 202 resume-initiated).
//
// See also: [Resume Sandbox API Reference]
//
// [Resume Sandbox API Reference]: https://docs.tensorlake.ai/api-reference/v2/sandboxes/resume
func (c *Client) ResumeSandbox(ctx context.Context, sandboxID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.sandboxAPIURL("/sandboxes/"+sandboxID+"/resume"), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = doSandboxAPI[struct{}](c, req, nil)
	return err
}
