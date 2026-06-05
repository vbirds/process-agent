# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`process-agent` is a single-package Go SDK (`package agent`, module `github.com/eric8810/process-agent`) that gives one unified interface for driving three different coding-agent CLIs: **Claude Code**, **Codex**, and **OpenCode**. Each CLI speaks a different wire protocol; this package hides those differences behind one `Backend` interface.

There is no `main` — this is a library, consumed via `agent.New(...)`.

## Commands

```bash
go build ./...        # compile
go test ./...         # run all tests
go test -run TestName # run a single test (e.g. go test -run TestNewBackendValidTypes)
go vet ./...          # vet
```

The tests are pure unit tests over the type layer (`agent_test.go`); they do **not** spawn real CLIs, so they run without `claude`/`codex`/`opencode` installed.

## Architecture

The contract lives in `agent.go` and is intentionally small:

- `Backend.Execute(ctx, prompt, opts) (*Session, error)` — the only interface method.
- A `Session` exposes two channels: `Messages` (streaming `Message` events, closed when done) and `Result` (exactly one final `Result`, then closed). Callers may ignore `Messages` entirely and just read `Result`.
- `New(agentType, cfg)` is the factory dispatching `"claude"`/`"codex"`/`"opencode"` to the three backend structs.

Each backend lives in its own file (`claude.go`, `codex.go`, `opencode.go`) and follows the **same execution skeleton**, which is the key pattern to preserve when editing:

1. Resolve the executable (default to the agent name), `exec.LookPath` it.
2. Apply timeout (default **20 minutes** when `opts.Timeout == 0`) via `context.WithTimeout`.
3. Spawn the CLI with stdout/stdin pipes; route stderr through `newLogWriter` (helpers.go) into the `slog.Logger`.
4. A reader goroutine scans stdout line-by-line (1MB initial / 10MB max scanner buffer) and translates protocol-specific events into the unified `Message` types.
5. A lifecycle goroutine drives the protocol, accumulates text into `Result.Output`, and sends the final `Result`. It owns closing `msgCh`/`resCh` and cancelling the context.

`Result.Status` is one of `"completed"`, `"failed"`, `"aborted"`, `"timeout"`, derived by inspecting `runCtx.Err()` (DeadlineExceeded → timeout, Canceled → aborted) plus the process exit error.

### Protocol differences (the reason this package exists)

- **Claude** (`claude.go`): one-shot CLI call `claude -p <prompt> --output-format stream-json --verbose --permission-mode bypassPermissions`. Reads JSON Lines; message `type` of `assistant`/`user`/`system`/`result`/`log`/`control_request`. Tool approval works by replying to `control_request` lines on stdin with a `control_response` (`behavior: "allow"`). Supports session resume via `--resume` and `MaxTurns`/`SystemPrompt` flags.
- **Codex** (`codex.go`): long-lived `codex app-server --listen stdio://` speaking **JSON-RPC 2.0**. The lifecycle is a handshake: `initialize` → `initialized` → `thread/start` → `turn/start`, then wait for turn completion. The `codexClient` struct is a full JSON-RPC client (request/response correlation by integer `id` via the `pending` map, notifications, server-request handling). It auto-detects two notification dialects at runtime (`notificationProtocol`: `"legacy"` `codex/event` envelopes vs. `"raw"` v2 `turn/*`+`item/*` notifications) — both paths must stay in sync when adding event handling.
- **OpenCode** (`opencode.go`): `opencode --json-stdio`, a Codex-compatible JSON-RPC variant; mirrors the codex structure.

Tool approval is **always auto-approved** across all backends — this SDK targets autonomous/daemon use with no human in the loop. Codex/OpenCode answer `requestApproval`/`*Approval` server requests with `{"decision":"accept"}`.

### Shared helpers (`helpers.go`)

- `trySend` — **non-blocking** channel send; drops the message if `Messages` is full. Streaming is best-effort; `Result.Output` is the source of truth for final text. Don't convert this to a blocking send (it would deadlock a slow/absent consumer).
- `buildEnv` — `os.Environ()` plus `Config.Env` overrides.
- `newLogWriter` / `detectCLIVersion`.

## Conventions

- Provider-specific request shapes use unexported types suffixed by provider (e.g. `claudeSDKMessage`, `codexClient`). Keep new protocol types in the relevant backend file, not `agent.go`.
- The public surface (`agent.go`) is provider-neutral. Anything provider-specific (flags, JSON-RPC methods, event names) belongs in the backend file and must be mapped to the shared `Message`/`Result` vocabulary before crossing the channel boundary.