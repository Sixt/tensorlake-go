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
	"fmt"
	"io"
	"net/http"
)

const (
	// SandboxAPIBaseURL is the default base URL for sandbox management operations.
	SandboxAPIBaseURL = "https://api.tensorlake.ai/sandboxes"
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
	Id                         string                       `json:"id"`
	Namespace                  string                       `json:"namespace"`
	Image                      string                       `json:"image,omitempty"`
	Status                     SandboxStatus                `json:"status"`
	PendingReason              string                       `json:"pending_reason,omitempty"`
	Outcome                    string                       `json:"outcome,omitempty"`
	CreatedAt                  int64                        `json:"created_at"`
	ContainerId                string                       `json:"container_id,omitempty"`
	ExecutorId                 string                       `json:"executor_id,omitempty"`
	Resources                  ContainerResourcesInfo       `json:"resources"`
	TimeoutSecs                int64                        `json:"timeout_secs"`
	SandboxURL                 string                       `json:"sandbox_url,omitempty"`
	PoolId                     string                       `json:"pool_id,omitempty"`
	NetworkPolicy              *SandboxNetworkAccessControl `json:"network_policy,omitempty"`
	AllowUnauthenticatedAccess bool                         `json:"allow_unauthenticated_access"`
	ExposedPorts               []int32                      `json:"exposed_ports,omitempty"`
	TemplateId                 string                       `json:"template_id,omitempty"`
	Name                       string                       `json:"name,omitempty"`
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

// sandboxAPIURL constructs a URL for sandbox management API calls.
// The path is appended to the sandbox API base URL which defaults to
// https://api.tensorlake.ai/sandboxes.
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
