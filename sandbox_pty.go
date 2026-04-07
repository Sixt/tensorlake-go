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
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/coder/websocket"
)

// PTY WebSocket binary frame opcodes.
const (
	ptyOpcodeData   byte = 0x00 // Terminal data (input or output)
	ptyOpcodeResize byte = 0x01 // Client → server: resize terminal
	ptyOpcodeReady  byte = 0x02 // Client → server: ready to receive output
	ptyOpcodeExit   byte = 0x03 // Server → client: process exited
)

// CreatePTYRequest holds options for creating a PTY session.
type CreatePTYRequest struct {
	// Command is the executable to run (e.g. "/bin/bash").
	//
	// Required.
	Command string `json:"command"`

	// Args are command-line arguments (e.g. ["-l"]).
	Args []string `json:"args,omitempty"`

	// Env sets environment variables for the session.
	Env map[string]string `json:"env,omitempty"`

	// WorkingDir is the initial working directory.
	WorkingDir string `json:"working_dir,omitempty"`

	// Rows is the terminal height. Default: 24. Clamped to 1..500.
	Rows int32 `json:"rows,omitempty"`

	// Cols is the terminal width. Default: 80. Clamped to 1..1000.
	Cols int32 `json:"cols,omitempty"`
}

// CreatePTYResponse represents the response from creating a PTY session.
type CreatePTYResponse struct {
	// SessionId is the unique PTY session identifier.
	SessionId string `json:"session_id"`

	// Token is used for WebSocket connection authentication.
	Token string `json:"token"`
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

// PTYListResponse represents the response from listing PTY sessions.
type PTYListResponse struct {
	Sessions []PTYSessionInfo `json:"sessions"`
}

// ResizePTYRequest holds the terminal dimensions for a resize operation.
type ResizePTYRequest struct {
	Rows int32 `json:"rows"`
	Cols int32 `json:"cols"`
}

// CreatePTY creates a new PTY session in a sandbox.
//
// Returns a session ID and token for WebSocket authentication.
// The maximum number of concurrent PTY sessions per sandbox is 64.
//
// See also: [Create PTY Session API Reference]
//
// [Create PTY Session API Reference]: https://docs.tensorlake.ai/api-reference/v2/pty/create
func (c *Client) CreatePTY(ctx context.Context, sandboxID string, in *CreatePTYRequest) (*CreatePTYResponse, error) {
	body, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	reqURL := c.sandboxProxyURL(sandboxID) + "/pty"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := doSandbox(c, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result CreatePTYResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}

// ListPTY lists all PTY sessions in a sandbox.
//
// The PTY token is not included in list responses.
//
// See also: [List PTY Sessions API Reference]
//
// [List PTY Sessions API Reference]: https://docs.tensorlake.ai/api-reference/v2/pty/list
func (c *Client) ListPTY(ctx context.Context, sandboxID string) (*PTYListResponse, error) {
	reqURL := c.sandboxProxyURL(sandboxID) + "/pty"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := doSandbox(c, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result PTYListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}

// GetPTY retrieves details for a specific PTY session.
//
// See also: [Get PTY Session API Reference]
//
// [Get PTY Session API Reference]: https://docs.tensorlake.ai/api-reference/v2/pty/get
func (c *Client) GetPTY(ctx context.Context, sandboxID, sessionID string) (*PTYSessionInfo, error) {
	reqURL := c.sandboxProxyURL(sandboxID) + "/pty/" + sessionID
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := doSandbox(c, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result PTYSessionInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}

// ResizePTY resizes a PTY session's terminal dimensions.
//
// Rows are clamped to 1..500, cols to 1..1000 server-side.
//
// See also: [Resize PTY Session API Reference]
//
// [Resize PTY Session API Reference]: https://docs.tensorlake.ai/api-reference/v2/pty/resize
func (c *Client) ResizePTY(ctx context.Context, sandboxID, sessionID string, in *ResizePTYRequest) error {
	body, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	reqURL := c.sandboxProxyURL(sandboxID) + "/pty/" + sessionID + "/resize"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := doSandbox(c, req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// KillPTY terminates a PTY session.
//
// The daemon sends SIGHUP initially, then escalates to SIGKILL
// if the session persists after a grace period.
//
// See also: [Kill PTY Session API Reference]
//
// [Kill PTY Session API Reference]: https://docs.tensorlake.ai/api-reference/v2/pty/kill
func (c *Client) KillPTY(ctx context.Context, sandboxID, sessionID string) error {
	reqURL := c.sandboxProxyURL(sandboxID) + "/pty/" + sessionID
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := doSandbox(c, req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// PTYConn represents an active WebSocket connection to a PTY session.
//
// PTYConn wraps a WebSocket connection and provides typed methods for
// the PTY binary protocol. After creating a PTYConn with [Client.ConnectPTY],
// the caller must call [PTYConn.Close] when done.
//
// The PTY WebSocket protocol uses binary frames with a single-byte opcode prefix:
//   - 0x00 Data: terminal I/O (both directions)
//   - 0x01 Resize: client → server terminal resize (uint16 BE cols + uint16 BE rows)
//   - 0x02 Ready: client → server readiness signal (must be sent first)
//   - 0x03 Exit: server → client process exit (int32 BE exit code)
type PTYConn struct {
	conn *websocket.Conn
	mu   sync.Mutex // protects writes
}

// ConnectPTY opens a WebSocket connection to a PTY session.
//
// The token is obtained from [Client.CreatePTY]. After connecting, the caller
// must call [PTYConn.Ready] to signal readiness before reading output.
//
// See also: [PTY WebSocket API Reference]
//
// [PTY WebSocket API Reference]: https://docs.tensorlake.ai/api-reference/v2/pty/websocket
func (c *Client) ConnectPTY(ctx context.Context, sandboxID, sessionID, token string) (*PTYConn, error) {
	proxyBase := c.sandboxProxyURL(sandboxID)
	// Convert https:// to wss://
	wsURL := "wss" + proxyBase[len("https"):] + "/pty/" + sessionID + "/ws"

	opts := &websocket.DialOptions{
		HTTPHeader: http.Header{
			"Authorization": {fmt.Sprintf("Bearer %s", c.apiKey)},
			"X-PTY-Token":  {token},
		},
	}

	conn, _, err := websocket.Dial(ctx, wsURL, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect PTY websocket: %w", err)
	}

	return &PTYConn{conn: conn}, nil
}

// ConnectPTYWithURL opens a WebSocket connection using an explicit base URL.
// This is primarily for testing with non-standard URLs.
func (c *Client) ConnectPTYWithURL(ctx context.Context, wsURL, token string) (*PTYConn, error) {
	// Ensure wss:// scheme
	u, err := url.Parse(wsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}
	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else if u.Scheme == "http" {
		u.Scheme = "ws"
	}

	opts := &websocket.DialOptions{
		HTTPHeader: http.Header{
			"Authorization": {fmt.Sprintf("Bearer %s", c.apiKey)},
			"X-PTY-Token":  {token},
		},
	}

	conn, _, err := websocket.Dial(ctx, u.String(), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect PTY websocket: %w", err)
	}

	return &PTYConn{conn: conn}, nil
}

// Ready sends the READY signal to the server, indicating the client is
// ready to receive terminal output. This must be called immediately after
// connecting, before reading any data.
//
// If Ready is not sent, the server buffers output up to 1 MB then disconnects.
func (pc *PTYConn) Ready(ctx context.Context) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	return pc.conn.Write(ctx, websocket.MessageBinary, []byte{ptyOpcodeReady})
}

// Write sends terminal input data to the PTY session.
func (pc *PTYConn) Write(ctx context.Context, data []byte) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	msg := make([]byte, 1+len(data))
	msg[0] = ptyOpcodeData
	copy(msg[1:], data)
	return pc.conn.Write(ctx, websocket.MessageBinary, msg)
}

// Resize sends a terminal resize notification to the PTY session.
func (pc *PTYConn) Resize(ctx context.Context, cols, rows uint16) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	msg := make([]byte, 5)
	msg[0] = ptyOpcodeResize
	binary.BigEndian.PutUint16(msg[1:3], cols)
	binary.BigEndian.PutUint16(msg[3:5], rows)
	return pc.conn.Write(ctx, websocket.MessageBinary, msg)
}

// PTYMessage represents a message received from the PTY WebSocket.
type PTYMessage struct {
	// Type is either PTYMessageData or PTYMessageExit.
	Type PTYMessageType

	// Data contains the terminal output bytes. Only set when Type is PTYMessageData.
	Data []byte

	// ExitCode contains the process exit code. Only set when Type is PTYMessageExit.
	ExitCode int32
}

// PTYMessageType distinguishes between data and exit messages.
type PTYMessageType int

const (
	// PTYMessageData indicates terminal output data.
	PTYMessageData PTYMessageType = iota
	// PTYMessageExit indicates the process has exited.
	PTYMessageExit
)

// Read reads the next message from the PTY session.
//
// Returns [PTYMessageData] for terminal output and [PTYMessageExit] when the
// process exits. After receiving an exit message, the WebSocket will be closed
// by the server.
func (pc *PTYConn) Read(ctx context.Context) (*PTYMessage, error) {
	_, data, err := pc.conn.Read(ctx)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, io.ErrUnexpectedEOF
	}

	switch data[0] {
	case ptyOpcodeData:
		return &PTYMessage{Type: PTYMessageData, Data: data[1:]}, nil
	case ptyOpcodeExit:
		if len(data) < 5 {
			return nil, fmt.Errorf("invalid exit frame: expected 5 bytes, got %d", len(data))
		}
		exitCode := int32(binary.BigEndian.Uint32(data[1:5]))
		return &PTYMessage{Type: PTYMessageExit, ExitCode: exitCode}, nil
	default:
		return nil, fmt.Errorf("unknown PTY opcode: 0x%02x", data[0])
	}
}

// Close closes the PTY WebSocket connection.
func (pc *PTYConn) Close() error {
	return pc.conn.Close(websocket.StatusNormalClosure, "")
}
