// Copyright 2025 SIXT SE
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
	"strings"

	"github.com/sixt/tensorlake-go/internal/sse"
)

// ParseResultUpdateFunc is a callback function that receives intermediate parse result updates
// during SSE streaming. It will be called for each SSE event received.
type ParseResultUpdateFunc func(name ParseEventName, result *ParseResult)

type GetParseResultOptions struct {
	withOptions bool

	// UseSSE enables Server-Sent Events (SSE) for streaming updates.
	// See also: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events
	useSSE bool
	// OnUpdate is a callback function that receives intermediate parse result updates
	// during SSE streaming. It will be called for each SSE event received.
	onUpdate ParseResultUpdateFunc
}

// GetParseResultOption is a function that configures the GetParseResultOptions.
type GetParseResultOption func(*GetParseResultOptions)

func WithOptions(enable bool) GetParseResultOption {
	return func(opts *GetParseResultOptions) {
		opts.withOptions = enable
	}
}

// WithSSE enables Server-Sent Events (SSE) for streaming updates.
func WithSSE(enable bool) GetParseResultOption {
	return func(opts *GetParseResultOptions) {
		opts.useSSE = enable
	}
}

// WithOnUpdate sets the callback function that receives intermediate parse result updates
// during SSE streaming. It will be called for each SSE event received.
func WithOnUpdate(onUpdate ParseResultUpdateFunc) GetParseResultOption {
	return func(opts *GetParseResultOptions) {
		opts.onUpdate = onUpdate
	}
}

// GetParseResult retrieves the result of a parse job.
// The response will include: 1) parsed content (markdown or pages);
// 2) structured extraction results (if schemas are provided during the parse request);
// 3) page classification results (if page classifications are provided during the parse request).
//
// When the job finishes successfully, the response will contain pages
// (chunks of the page) chunks (text chunks extracted from the document),
// structured data (every schema_name provided in the parse request as a key).
//
// See also: [Get Parse Result API Reference]
//
// [Get Parse Result API Reference]: https://docs.tensorlake.ai/api-reference/v2/parse/get
func (c *Client) GetParseResult(ctx context.Context, parseId string, opts ...GetParseResultOption) (*ParseResult, error) {
	o := &GetParseResultOptions{
		withOptions: false,
		useSSE:      false,
		onUpdate:    nil,
	}
	for _, opt := range opts {
		opt(o)
	}

	reqURL := fmt.Sprintf("%s/parse/%s?with_options=%t", c.baseURL, parseId, o.withOptions)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if o.useSSE {
		return c.handleSSEResponse(req, o.onUpdate)
	}

	return do(c, req, func(r io.Reader) (*ParseResult, error) {
		var result ParseResult
		if err := json.NewDecoder(r).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return &result, nil
	})
}

// ParseEventName is the name of the SSE event.
type ParseEventName string

// The possible SSE events.
// See also: https://github.com/tensorlakeai/tensorlake/blob/main/src/tensorlake/documentai/_parse.py#L499
const (
	SSEEventParseQueued ParseEventName = "parse_queued"
	SSEEventParseUpdate ParseEventName = "parse_update"
	SSEEventParseDone   ParseEventName = "parse_done"
	SSEEventParseFailed ParseEventName = "parse_failed"
)

func (c *Client) handleSSEResponse(req *http.Request, onUpdate ParseResultUpdateFunc) (*ParseResult, error) {
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute SSE request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errRes ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errRes); err != nil {
			return nil, fmt.Errorf("failed to decode SSE error response (%d): %w", resp.StatusCode, err)
		}
		return nil, &errRes
	}

	// Check if response is actually SSE
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") {
		// Fall back to regular JSON parsing
		var result ParseResult
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode SSE response: %w", err)
		}
		return &result, nil
	}

	var lastEvent string
	var eventCount int
	for ev, err := range sse.ScanEvents(resp.Body) {
		if err != nil {
			return nil, fmt.Errorf("failed to scan SSE events: %w", err)
		}

		lastEvent = ev.Name()
		eventCount++

		// Unmarshal the event data into a ParseResult.
		var result ParseResult
		if err := json.Unmarshal(ev.Data(), &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal SSE event: %w", err)
		}

		switch ev.Name() {
		case string(SSEEventParseQueued), string(SSEEventParseUpdate):
			if onUpdate != nil {
				onUpdate(ParseEventName(ev.Name()), &result)
			}
			continue

		case string(SSEEventParseDone):
			if onUpdate != nil {
				onUpdate(ParseEventName(ev.Name()), &result)
			}
			return &result, nil

		case string(SSEEventParseFailed):
			if onUpdate != nil {
				onUpdate(ParseEventName(ev.Name()), &result)
			}
			return nil, fmt.Errorf("failed to parse result: %s", result.Error)
		default:
			return nil, fmt.Errorf("unknown SSE event: %s", ev.Name())
		}
	}

	if eventCount == 0 {
		return nil, fmt.Errorf("SSE stream closed without sending any events")
	}
	return nil, fmt.Errorf("SSE stream ended after %d event(s), last event %q, without a terminal event (parse_done or parse_failed)", eventCount, lastEvent)
}
