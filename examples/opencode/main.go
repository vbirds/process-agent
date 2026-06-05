// Command opencode-example runs a single prompt through the OpenCode backend
// and prints the streamed messages plus the final result.
//
// Usage:
//
//	go run ./examples/opencode "Write a hello world in Go"
//
// Requires the `opencode` CLI to be installed and on PATH (or set OPENCODE_PATH).
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	agent "github.com/eric8810/process-agent"
)

func main() {
	prompt := "List the files in the current directory and summarize what this project does."
	if len(os.Args) > 1 {
		prompt = os.Args[1]
	}

	// Create the OpenCode backend. ExecutablePath defaults to "opencode" when empty.
	backend, err := agent.New("opencode", agent.Config{
		ExecutablePath: os.Getenv("OPENCODE_PATH"),
		Logger:         slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})),
	})
	if err != nil {
		log.Fatalf("create backend: %v", err)
	}

	ctx := context.Background()

	session, err := backend.Execute(ctx, prompt, agent.ExecOptions{
		Cwd:     ".",
		Timeout: 5 * time.Minute,
	})
	if err != nil {
		log.Fatalf("execute: %v", err)
	}

	// Stream events as the agent works.
	for msg := range session.Messages {
		switch msg.Type {
		case agent.MessageText:
			fmt.Print(msg.Content)
		case agent.MessageThinking:
			fmt.Printf("\n[thinking] %s\n", truncate(msg.Content, 200))
		case agent.MessageToolUse:
			fmt.Printf("\n[tool ▶ %s] %v\n", msg.Tool, msg.Input)
		case agent.MessageToolResult:
			fmt.Printf("[tool ✓ %s] %s\n", msg.Tool, truncate(msg.Output, 200))
		case agent.MessageStatus:
			fmt.Printf("\n[status: %s]\n", msg.Status)
		case agent.MessageError:
			fmt.Printf("\n[error] %s\n", msg.Content)
		case agent.MessageLog:
			fmt.Printf("[log:%s] %s\n", msg.Level, msg.Content)
		}
	}

	// Exactly one final result is delivered after Messages closes.
	result := <-session.Result
	fmt.Printf("\n\n──────── result ────────\n")
	fmt.Printf("Status:   %s\n", result.Status)
	fmt.Printf("Duration: %dms\n", result.DurationMs)
	if result.Error != "" {
		fmt.Printf("Error:    %s\n", result.Error)
	}

	if result.Status != "completed" {
		os.Exit(1)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
