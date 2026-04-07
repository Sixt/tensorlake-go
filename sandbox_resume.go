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

// ResumeSandbox resumes a suspended sandbox.
//
// Returns nil on success (both 200 already-running and 202 resume-initiated).
//
// See also: [Resume Sandbox API Reference]
//
// [Resume Sandbox API Reference]: https://docs.tensorlake.ai/api-reference/v2/sandboxes/resume
func (c *Client) ResumeSandbox(ctx context.Context, sandboxID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.sandboxAPIURL("/"+sandboxID+"/resume"), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = doSandboxAPI[struct{}](c, req, nil)
	return err
}
