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
	"encoding/binary"
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
	u, err := url.Parse(wsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}
	switch u.Scheme {
	case "https":
		u.Scheme = "wss"
	case "http":
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
