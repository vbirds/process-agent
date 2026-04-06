package agent

import (
	"context"
	"testing"
	"time"
)

func TestExecOptionsDefaults(t *testing.T) {
	opts := ExecOptions{}
	if opts.Timeout != 0 {
		t.Errorf("expected default timeout 0, got %v", opts.Timeout)
	}
	if opts.MaxTurns != 0 {
		t.Errorf("expected default max turns 0, got %d", opts.MaxTurns)
	}
}

func TestMessageTypes(t *testing.T) {
	types := []MessageType{
		MessageText,
		MessageThinking,
		MessageToolUse,
		MessageToolResult,
		MessageStatus,
		MessageError,
		MessageLog,
	}

	expected := []string{
		"text", "thinking", "tool-use", "tool-result", "status", "error", "log",
	}

	for i, mt := range types {
		if string(mt) != expected[i] {
			t.Errorf("expected message type %s, got %s", expected[i], mt)
		}
	}
}

func TestMessageFields(t *testing.T) {
	msg := Message{
		Type:    MessageText,
		Content: "Hello world",
	}
	if msg.Type != MessageText {
		t.Errorf("expected type text, got %s", msg.Type)
	}
	if msg.Content != "Hello world" {
		t.Errorf("expected content 'Hello world', got %s", msg.Content)
	}
}

func TestMessageToolUse(t *testing.T) {
	input := map[string]any{"command": "ls -la"}
	msg := Message{
		Type:   MessageToolUse,
		Tool:   "bash",
		CallID: "call-123",
		Input:  input,
	}
	if msg.Tool != "bash" {
		t.Errorf("expected tool 'bash', got %s", msg.Tool)
	}
	if msg.CallID != "call-123" {
		t.Errorf("expected call_id 'call-123', got %s", msg.CallID)
	}
	if msg.Input["command"] != "ls -la" {
		t.Errorf("expected command 'ls -la', got %v", msg.Input["command"])
	}
}

func TestMessageToolResult(t *testing.T) {
	msg := Message{
		Type:   MessageToolResult,
		Tool:   "bash",
		CallID: "call-123",
		Output: "file1.txt\nfile2.txt",
	}
	if msg.Output != "file1.txt\nfile2.txt" {
		t.Errorf("expected output, got %s", msg.Output)
	}
}

func TestResultFields(t *testing.T) {
	result := Result{
		Status:     "completed",
		Output:     "Task done",
		DurationMs: 1000,
		SessionID:  "session-abc",
	}
	if result.Status != "completed" {
		t.Errorf("expected status 'completed', got %s", result.Status)
	}
	if result.DurationMs != 1000 {
		t.Errorf("expected duration 1000ms, got %d", result.DurationMs)
	}
}

func TestConfigFields(t *testing.T) {
	cfg := Config{
		ExecutablePath: "/usr/bin/claude",
		Env:           map[string]string{"KEY": "value"},
	}
	if cfg.ExecutablePath != "/usr/bin/claude" {
		t.Errorf("expected executable path, got %s", cfg.ExecutablePath)
	}
	if cfg.Env["KEY"] != "value" {
		t.Errorf("expected env value, got %s", cfg.Env["KEY"])
	}
}

func TestNewBackendInvalidType(t *testing.T) {
	_, err := New("invalid-type", Config{})
	if err == nil {
		t.Error("expected error for invalid backend type")
	}
}

func TestNewBackendValidTypes(t *testing.T) {
	types := []string{"claude", "codex", "opencode"}
	for _, agentType := range types {
		backend, err := New(agentType, Config{})
		if err != nil {
			t.Errorf("expected no error for type %s, got %v", agentType, err)
		}
		if backend == nil {
			t.Errorf("expected backend for type %s", agentType)
		}
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(time.Second):
		t.Error("context should have been cancelled")
	}
}
