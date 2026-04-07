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
	"iter"
	"net/http"
)

// FollowProcessStdout opens an SSE stream that replays captured stdout
// then streams live stdout lines until the process exits.
//
// The returned iterator yields [ProcessOutputEvent] for each line.
// The Stream field is not set (stdout-only endpoint).
// Iteration ends when the server sends an "eof" event.
//
// See also: [Follow Process Stdout API Reference]
//
// [Follow Process Stdout API Reference]: https://docs.tensorlake.ai/api-reference/v2/processes/follow-stdout
func (c *Client) FollowProcessStdout(ctx context.Context, sandboxID string, pid int32) iter.Seq2[ProcessOutputEvent, error] {
	reqURL := fmt.Sprintf("%s/processes/%d/stdout/follow", c.sandboxProxyURL(sandboxID), pid)
	return c.followProcessOutput(ctx, reqURL)
}

func (c *Client) followProcessStream(ctx context.Context, reqURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "text/event-stream")
	return doSandbox(c, req)
}
