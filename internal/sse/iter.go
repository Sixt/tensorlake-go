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

package sse

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"iter"
	"strings"
)

// An Event is a server-sent event.
//
// See https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events#fields.
type Event struct {
	id    string // the "id" field
	name  string // the "event" field
	data  []byte // the "data" field
	retry string // the "retry" field
}

// Empty returns true if the Event is empty.
func (e Event) Empty() bool {
	return e.name == "" && e.id == "" && len(e.data) == 0 && e.retry == ""
}

// Name returns the name of the event.
func (e Event) Name() string {
	return e.name
}

// Id returns the id of the event.
func (e Event) Id() string {
	return e.id
}

// Data returns the data of the event.
func (e Event) Data() []byte {
	return e.data
}

// Retry returns the retry of the event.
func (e Event) Retry() string {
	return e.retry
}

// ScanEvents iterates SSE events in the given io.Reader. The iterated error is
// terminal: if encountered, the stream is corrupt or broken and should no longer
// be used.
func ScanEvents(r io.Reader) iter.Seq2[Event, error] {
	// This code is modified from https://github.com/modelcontextprotocol/go-sdk/blob/411d5a049ae1006cf01ab9f5ac49d249ab3e8e2b/mcp/event.go#L69
	//
	// Original license:
	//
	// Copyright 2025 The Go MCP SDK Authors. All rights reserved.
	// Use of this source code is governed by an MIT-style
	// license that can be found in the LICENSE file.

	s := bufio.NewScanner(r)
	const maxTokenSize = 1 * 1024 * 1024 // 1 MiB max line size
	s.Buffer(nil, maxTokenSize)

	var (
		idKey    = []byte("id")
		eventKey = []byte("event")
		dataKey  = []byte("data")
		retryKey = []byte("retry")
	)

	return func(yield func(Event, error) bool) {
		// Reads and parses events from the stream following the SSE protocol specification:
		// See: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events#examples
		//
		//   - Each line follows the `key: value` format.
		//   - Multiple consecutive `data:` lines are combined using newline separation.
		//   - Only the 'event', 'id', 'data', and 'retry' fields are processed; all other fields are ignored.
		//   - Lines starting with ":" are treated as comments and ignored.
		//   - An event record ends when a blank line (\n\n) is encountered.
		var (
			evt     Event
			dataBuf *bytes.Buffer // if non-nil, preceding field was also data
		)
		flushData := func() {
			if dataBuf != nil {
				evt.data = dataBuf.Bytes()
				dataBuf = nil
			}
		}
		for s.Scan() {
			line := s.Bytes()
			if len(line) == 0 {
				flushData()
				// \n\n is the record delimiter
				if !evt.Empty() && !yield(evt, nil) {
					return
				}
				evt = Event{}
				continue
			}

			before, after, found := bytes.Cut(line, []byte{':'})
			if !found {
				yield(Event{}, fmt.Errorf("malformed line in SSE stream: %q", string(line)))
				return
			}
			if !bytes.Equal(before, dataKey) {
				flushData()
			}
			switch {
			case bytes.Equal(before, eventKey):
				evt.name = strings.TrimSpace(string(after))
			case bytes.Equal(before, idKey):
				evt.id = strings.TrimSpace(string(after))
			case bytes.Equal(before, retryKey):
				evt.retry = strings.TrimSpace(string(after))
			case bytes.Equal(before, dataKey):
				data := bytes.TrimSpace(after)
				if dataBuf != nil {
					dataBuf.WriteByte('\n')
					dataBuf.Write(data)
				} else {
					dataBuf = new(bytes.Buffer)
					dataBuf.Write(data)
				}
			}
		}
		if err := s.Err(); err != nil {
			if errors.Is(err, bufio.ErrTooLong) {
				err = fmt.Errorf("event exceeded max line length of %d", maxTokenSize)
			}
			if !yield(Event{}, err) {
				return
			}
		}
		flushData()
		if !evt.Empty() {
			yield(evt, nil)
		}
	}
}
