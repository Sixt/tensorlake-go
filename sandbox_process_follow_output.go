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
	"iter"

	"github.com/sixt/tensorlake-go/internal/sse"
)

// FollowProcessOutput opens an SSE stream that replays captured output
// (stdout + stderr merged) then streams live output lines until the process exits.
//
// The returned iterator yields [ProcessOutputEvent] for each line.
// The Stream field is set to "stdout" or "stderr" to identify the source.
// Iteration ends when the server sends an "eof" event.
//
// See also: [Follow Process Output API Reference]
//
// [Follow Process Output API Reference]: https://docs.tensorlake.ai/api-reference/v2/processes/follow-output
func (c *Client) FollowProcessOutput(ctx context.Context, sandboxID string, pid int32) iter.Seq2[ProcessOutputEvent, error] {
	reqURL := fmt.Sprintf("%s/processes/%d/output/follow", c.sandboxProxyURL(sandboxID), pid)
	return c.followProcessOutput(ctx, reqURL)
}

// followProcessOutput is the shared implementation for all three follow endpoints.
func (c *Client) followProcessOutput(ctx context.Context, reqURL string) iter.Seq2[ProcessOutputEvent, error] {
	return func(yield func(ProcessOutputEvent, error) bool) {
		resp, err := c.followProcessStream(ctx, reqURL)
		if err != nil {
			yield(ProcessOutputEvent{}, err)
			return
		}
		defer resp.Body.Close()

		for evt, err := range sse.ScanEvents(resp.Body) {
			if err != nil {
				yield(ProcessOutputEvent{}, fmt.Errorf("failed to scan SSE events: %w", err))
				return
			}

			switch evt.Name() {
			case "eof":
				return
			case "output":
				var out ProcessOutputEvent
				if err := json.Unmarshal(evt.Data(), &out); err != nil {
					yield(ProcessOutputEvent{}, fmt.Errorf("failed to decode output event: %w", err))
					return
				}
				if !yield(out, nil) {
					return
				}
			}
		}
	}
}
