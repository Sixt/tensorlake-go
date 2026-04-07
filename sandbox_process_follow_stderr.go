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
)

// FollowProcessStderr opens an SSE stream that replays captured stderr
// then streams live stderr lines until the process exits.
//
// The returned iterator yields [ProcessOutputEvent] for each line.
// The Stream field is not set (stderr-only endpoint).
// Iteration ends when the server sends an "eof" event.
//
// See also: [Follow Process Stderr API Reference]
//
// [Follow Process Stderr API Reference]: https://docs.tensorlake.ai/api-reference/v2/processes/follow-stderr
func (c *Client) FollowProcessStderr(ctx context.Context, sandboxID string, pid int32) iter.Seq2[ProcessOutputEvent, error] {
	reqURL := fmt.Sprintf("%s/processes/%d/stderr/follow", c.sandboxProxyURL(sandboxID), pid)
	return c.followProcessOutput(ctx, reqURL)
}
