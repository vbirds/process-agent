# process-agent

A unified Go SDK for executing prompts via coding agents (Claude Code, Codex, OpenCode).

## Overview

Each coding agent CLI has its own communication protocol:

- **Claude Code**: JSON Lines streaming + control protocol
- **Codex**: JSON-RPC 2.0 over stdio
- **OpenCode**: JSON-RPC 2.0 (Codex-compatible)

`process-agent` provides a unified interface that abstracts away these protocol differences, letting you integrate multiple agent backends with a single API.

## Installation

```bash
go get github.com/eric8810/process-agent
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    agent "github.com/eric8810/process-agent"
)

func main() {
    ctx := context.Background()

    // Create a backend
    backend, err := agent.New("claude", agent.Config{
        // ExecutablePath: "/path/to/claude", // optional, defaults to "claude"
        Logger: nil, // uses slog.Default()
    })
    if err != nil {
        log.Fatal(err)
    }

    // Execute a prompt
    session, err := backend.Execute(ctx, "Write a hello world in Go", agent.ExecOptions{
        Cwd:     ".",
        Timeout: 5 * time.Minute,
    })
    if err != nil {
        log.Fatal(err)
    }

    // Stream messages (optional)
    for msg := range session.Messages {
        switch msg.Type {
        case agent.MessageText:
            fmt.Print(msg.Content)
        case agent.MessageToolUse:
            fmt.Printf("\n[tool: %s]\n", msg.Tool)
        case agent.MessageToolResult:
            fmt.Printf("[result: %s]\n", truncate(msg.Output, 100))
        }
    }

    // Get final result
    result := <-session.Result
    fmt.Printf("\n\nStatus: %s\n", result.Status)
    fmt.Printf("Duration: %dms\n", result.DurationMs)
}

func truncate(s string, max int) string {
    if len(s) <= max {
        return s
    }
    return s[:max] + "..."
}
```

## Supported Backends

### Claude Code

```go
backend, _ := agent.New("claude", agent.Config{
    ExecutablePath: "claude", // optional
})
```

**Options**:
- `opts.Model` - Model to use (e.g., "claude-sonnet-4-6")
- `opts.ResumeSessionID` - Resume a previous session
- `opts.MaxTurns` - Maximum number of turns

### Codex

```go
backend, _ := agent.New("codex", agent.Config{
    ExecutablePath: "codex", // optional
})
```

**Options**:
- `opts.Model` - Model to use
- `opts.SystemPrompt` - Developer instructions

### OpenCode

```go
backend, _ := agent.New("opencode", agent.Config{
    ExecutablePath: "opencode", // optional
})
```

## Message Types

```go
const (
    MessageText       // Text output from the agent
    MessageThinking   // Agent's thinking process
    MessageToolUse    // Tool call initiated
    MessageToolResult // Tool call completed
    MessageStatus     // Agent status update
    MessageError      // Error occurred
    MessageLog        // Log message
)
```

## ExecOptions

```go
type ExecOptions struct {
    Cwd             string        // Working directory
    Model           string        // Model to use (provider-specific)
    SystemPrompt    string        // Additional system prompt
    MaxTurns        int           // Maximum turns (0 = unlimited)
    Timeout         time.Duration // Execution timeout (0 = 20 minutes)
    ResumeSessionID string        // Resume previous session (if supported)
}
```

## Result

```go
type Result struct {
    Status     string // "completed", "failed", "aborted", "timeout"
    Output     string // Accumulated text output
    Error      string // Error message if failed
    DurationMs int64  // Execution duration in milliseconds
    SessionID  string // Session ID for resumption
}
```

## Auto-Approval

All backends automatically approve tool calls and file operations. This is designed for autonomous/daemon mode where human interaction is not expected.

## License

MIT
