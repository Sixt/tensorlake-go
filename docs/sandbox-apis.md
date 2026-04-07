# Sandbox APIs

The sandbox APIs allow you to create, manage, and interact with cloud sandboxes. Sandboxes provide isolated environments for running processes, managing files, and connecting interactive terminal sessions.

## Sandbox Management

### Create a Sandbox

```go
resp, err := c.CreateSandbox(ctx, &tensorlake.CreateSandboxRequest{
    Name:        "my-sandbox",           // Optional: enables suspend/resume
    TimeoutSecs: ptr(int64(300)),        // Auto-terminate after 5 minutes
    Resources: &tensorlake.SandboxResourceOverrides{
        CPUs:     2.0,
        MemoryMB: 4096,
    },
    Network: &tensorlake.SandboxNetworkAccessControl{
        AllowInternetAccess: true,
    },
})
// resp.SandboxId, resp.Status
```

### List Sandboxes

```go
resp, err := c.ListSandboxes(ctx, &tensorlake.ListSandboxesRequest{
    Limit:  50,
    Status: "running",  // Filter by status
})
for _, sb := range resp.Sandboxes {
    fmt.Printf("%s: %s\n", sb.Id, sb.Status)
}
```

### Get Sandbox Details

```go
info, err := c.GetSandbox(ctx, sandboxID)
// info.Status, info.Resources, info.Name, etc.
```

### Update Sandbox

```go
info, err := c.UpdateSandbox(ctx, sandboxID, &tensorlake.UpdateSandboxRequest{
    ExposedPorts: []int32{8080, 3000},
})
```

### Delete (Terminate) Sandbox

```go
err := c.DeleteSandbox(ctx, sandboxID)
// Idempotent: terminating an already-terminated sandbox succeeds
```

### Suspend and Resume

Only named sandboxes support suspend/resume.

```go
err := c.SuspendSandbox(ctx, sandboxID)
// Wait for suspended status...
err = c.ResumeSandbox(ctx, sandboxID)
```

### Snapshot and Restore

```go
// Create snapshot
snap, err := c.SnapshotSandbox(ctx, sandboxID, &tensorlake.SnapshotSandboxRequest{
    SnapshotContentMode: tensorlake.SnapshotContentModeFull,
})

// Restore from snapshot
restored, err := c.CreateSandbox(ctx, &tensorlake.CreateSandboxRequest{
    SnapshotId: snap.SnapshotId,
})
```

### Sandbox Status

| Status | Description |
|--------|-------------|
| `pending` | Sandbox is being scheduled/provisioned |
| `running` | Sandbox is ready for use |
| `snapshotting` | Snapshot is being created |
| `suspending` | Sandbox is being suspended |
| `suspended` | Sandbox is suspended (can be resumed) |
| `terminated` | Sandbox has been terminated |

## File Operations

File operations use the sandbox proxy URL (`{id}.sandbox.tensorlake.ai`).

### Write a File

```go
content := bytes.NewReader([]byte("hello world"))
err := c.WriteSandboxFile(ctx, sandboxID, "/workspace/hello.txt", content)
// Parent directories are created automatically
```

### Read a File

```go
data, err := c.ReadSandboxFile(ctx, sandboxID, "/workspace/hello.txt")
// data is []byte of raw file content
```

### List a Directory

```go
resp, err := c.ListSandboxDirectory(ctx, sandboxID, "/workspace")
for _, entry := range resp.Entries {
    if entry.IsDir {
        fmt.Printf("[dir]  %s\n", entry.Name)
    } else {
        fmt.Printf("[file] %s (%d bytes)\n", entry.Name, *entry.Size)
    }
}
```

### Delete a File

```go
err := c.DeleteSandboxFile(ctx, sandboxID, "/workspace/hello.txt")
```

## PTY Sessions

PTY sessions provide interactive terminal access to a sandbox via WebSocket.

### Create a PTY Session

```go
pty, err := c.CreatePTY(ctx, sandboxID, &tensorlake.CreatePTYRequest{
    Command:    "/bin/sh",
    Env:        map[string]string{"TERM": "xterm-256color"},
    WorkingDir: "/workspace",
    Rows:       24,
    Cols:       80,
})
// pty.SessionId, pty.Token
```

### Connect via WebSocket

```go
conn, err := c.ConnectPTY(ctx, sandboxID, pty.SessionId, pty.Token)
defer conn.Close()

// Must send Ready before reading output
conn.Ready(ctx)

// Send input
conn.Write(ctx, []byte("ls -la\n"))

// Read output
msg, err := conn.Read(ctx)
switch msg.Type {
case tensorlake.PTYMessageData:
    fmt.Print(string(msg.Data))
case tensorlake.PTYMessageExit:
    fmt.Printf("Process exited with code %d\n", msg.ExitCode)
}

// Resize terminal
conn.Resize(ctx, 120, 40)
```

### List / Get / Resize / Kill PTY Sessions

```go
// List all sessions
list, err := c.ListPTY(ctx, sandboxID)

// Get session details
info, err := c.GetPTY(ctx, sandboxID, sessionID)
// info.PID, info.IsAlive, info.Rows, info.Cols

// Resize via REST
err = c.ResizePTY(ctx, sandboxID, sessionID, &tensorlake.ResizePTYRequest{
    Rows: 40, Cols: 120,
})

// Kill session
err = c.KillPTY(ctx, sandboxID, sessionID)
```

## Process Management

Process APIs let you start, monitor, and control processes inside a sandbox.

### Start a Process

```go
proc, err := c.StartProcess(ctx, sandboxID, &tensorlake.StartProcessRequest{
    Command:    "python",
    Args:       []string{"-c", "print('hello')"},
    Env:        map[string]string{"PYTHONPATH": "/app"},
    WorkingDir: "/workspace",
    StdinMode:  tensorlake.StdinModePipe,    // "closed" (default) or "pipe"
    StdoutMode: tensorlake.OutputModeCapture, // "capture" (default) or "discard"
    StderrMode: tensorlake.OutputModeCapture,
})
// proc.PID, proc.Status, proc.StdinWritable
```

### List / Get Processes

```go
list, err := c.ListProcesses(ctx, sandboxID)

info, err := c.GetProcess(ctx, sandboxID, pid)
// info.Status: "running", "exited", or "signaled"
// info.ExitCode, info.Signal (nullable)
```

### Send Signal / Kill

```go
// Send SIGTERM
err := c.SignalProcess(ctx, sandboxID, pid, &tensorlake.SignalProcessRequest{
    Signal: 15,
})

// Kill (SIGKILL)
err := c.KillProcess(ctx, sandboxID, pid)
```

### Stdin Pipe

```go
// Write data to stdin (process must be started with StdinMode: "pipe")
err := c.WriteProcessStdin(ctx, sandboxID, pid, bytes.NewReader([]byte("input data\n")))

// Close stdin (sends EOF)
err := c.CloseProcessStdin(ctx, sandboxID, pid)
```

### Read Captured Output

```go
stdout, err := c.GetProcessStdout(ctx, sandboxID, pid)
stderr, err := c.GetProcessStderr(ctx, sandboxID, pid)
output, err := c.GetProcessOutput(ctx, sandboxID, pid) // merged

// All return: {PID, Lines []string, LineCount int32}
for _, line := range stdout.Lines {
    fmt.Println(line)
}
```

### Follow Output via SSE

Stream output in real-time using Server-Sent Events:

```go
// Follow stdout only
for evt, err := range c.FollowProcessStdout(ctx, sandboxID, pid) {
    if err != nil { break }
    fmt.Printf("%s\n", evt.Line)
}

// Follow stderr only
for evt, err := range c.FollowProcessStderr(ctx, sandboxID, pid) {
    if err != nil { break }
    fmt.Fprintf(os.Stderr, "%s\n", evt.Line)
}

// Follow merged output (with stream tags)
for evt, err := range c.FollowProcessOutput(ctx, sandboxID, pid) {
    if err != nil { break }
    fmt.Printf("[%s] %s\n", evt.Stream, evt.Line) // evt.Stream is "stdout" or "stderr"
}
```

The follow endpoints replay all previously captured output first, then stream live output until the process exits (signaled by an `eof` SSE event).

## Configuration

All sandbox URLs are configurable:

```go
c := tensorlake.NewClient(
    tensorlake.WithAPIKey("your-key"),
    // Management API (create, list, get, update, delete, snapshot, suspend, resume)
    tensorlake.WithSandboxAPIBaseURL("https://api-tensorlake.example.com/sandboxes"),
    // Proxy API (files, PTY, processes)
    tensorlake.WithSandboxProxyBaseURL("https://sandbox-tensorlake.example.com"),
)
```

| Option | Default | Used by |
|--------|---------|---------|
| `WithSandboxAPIBaseURL` | `https://api.tensorlake.ai/sandboxes` | Management APIs |
| `WithSandboxProxyBaseURL` | `https://sandbox.tensorlake.ai` | File, PTY, Process APIs |

## Error Handling

Sandbox proxy errors return `*SandboxProxyError`:

```go
_, err := c.ReadSandboxFile(ctx, sandboxID, "/nonexistent")
if err != nil {
    var sandboxErr *tensorlake.SandboxProxyError
    if errors.As(err, &sandboxErr) {
        fmt.Printf("Error: %s (code: %s)\n", sandboxErr.Err, sandboxErr.Code)
    }
}
```

Management API errors return plain text with HTTP status codes.
