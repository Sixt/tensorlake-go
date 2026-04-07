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

// ProcessStatus represents the current state of a sandbox process.
type ProcessStatus string

const (
	ProcessStatusRunning  ProcessStatus = "running"
	ProcessStatusExited   ProcessStatus = "exited"
	ProcessStatusSignaled ProcessStatus = "signaled"
)

// StdinMode determines how stdin is handled for a process.
type StdinMode string

const (
	StdinModeClosed StdinMode = "closed"
	StdinModePipe   StdinMode = "pipe"
)

// OutputMode determines how stdout/stderr is handled for a process.
type OutputMode string

const (
	OutputModeCapture OutputMode = "capture"
	OutputModeDiscard OutputMode = "discard"
)

// ProcessInfo represents metadata about a sandbox process.
type ProcessInfo struct {
	PID           int32         `json:"pid"`
	Status        ProcessStatus `json:"status"`
	ExitCode      *int32        `json:"exit_code,omitempty"`
	Signal        *int32        `json:"signal,omitempty"`
	StdinWritable bool          `json:"stdin_writable"`
	Command       string        `json:"command"`
	Args          []string      `json:"args"`
	StartedAt     int64         `json:"started_at"`
	EndedAt       *int64        `json:"ended_at,omitempty"`
}

// ProcessOutputResponse represents captured output lines from a process.
type ProcessOutputResponse struct {
	PID       int32    `json:"pid"`
	Lines     []string `json:"lines"`
	LineCount int32    `json:"line_count"`
}

// ProcessOutputEvent represents a single line of output from a follow stream.
type ProcessOutputEvent struct {
	Line      string `json:"line"`
	Timestamp int64  `json:"timestamp"`
	Stream    string `json:"stream,omitempty"` // "stdout" or "stderr", only present in follow-output
}
