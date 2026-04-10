# MVP — first version

This document scopes the **first shippable version** of CentrAI Agent: a Go library (and thin CLI) that runs a **single-agent** control loop with **native tools**, **HTTP-backed models**, and **durable-enough** session persistence for development and early integration. **Every model turn runs over a streaming API** (token/chunk events); there is no MVP requirement for a separate non-streaming completion-only path. It aligns with [8. architecture](8.%20architecture.md) and [9. code-structure](9.%20code-structure.md); detailed behavior remains in those docs and the foundation series ([1. agents](1.%20agents.md)–[7. skills](7.%20skills.md)).

**Status:** The repository includes a **Go implementation** aligned with this scope (`go.mod`, `internal/*`, `cmd/centrai`). Treat the sections below as **product and engineering scope** for v1; use the acceptance checklist to track release readiness.

---

## Goals

| Goal | What “done” means for v1 |
|------|---------------------------|
| **Runnable loop** | One **run** per user turn: load session → build prompt → **stream** from model → reassemble assistant text and/or **tool calls** from stream → execute tools → persist → repeat until finish or **max steps**, with **cancellation** via `context.Context`. |
| **Streaming model** | A **`model.Client`** (or equivalent) that uses the provider’s **streaming** chat API (e.g. SSE/chunked responses); it exposes **chunk or token events** to the orchestrator/host and yields a **final** assistant message (and tool calls when present) for the loop. **One** concrete provider in-tree (e.g. OpenAI-compatible streaming chat) is enough for MVP. **Batch/non-streaming HTTP completion** is not required for v1. |
| **Native tools** | **`tool.Registry`**: register tools with JSON Schema (or subset), validate arguments, execute with **per-call timeout**, append tool results, deterministic ordering when multiple calls exist. |
| **Session + history** | **`session` domain types** (`Message`, roles, tool call ids) and **`session.Store`** implemented at least by an **in-memory** backend; optional **SQLite** adapter if persistence is required for the first release. |
| **Prompt assembly** | **`prompt` builder** merges system instructions, truncated history, and tool definitions into provider-ready payloads—no full skill pipeline required for MVP (see Out of scope). |
| **Host wiring** | **`cmd/centrai`** (or equivalent) reads **env/flags**, constructs dependencies, and runs a minimal path (e.g. one-shot message or REPL)—**no** business rules in `main` beyond wiring. |
| **Quality bar** | `go test ./...`, `gofmt`/`goimports`, and baseline lint ([9. code-structure §11](9.%20code-structure.md#11-tooling-recommended)); unit tests with **`httptest`** for HTTP adapters, **fakes** for `Store` / model where useful. |

---

## Out of scope for v1

The following are **explicitly not required** for the first version. They stay in the architecture as **target** behavior and can follow in later releases.

| Area | Deferred |
|------|----------|
| **MCP** | No `internal/mcp` client or remote tool discovery in v1 ([6. mcps](6.%20mcps.md)). |
| **Skills loader** | No filesystem or registry **skill** resolution in v1; callers may pass **static system text** into the prompt builder instead ([7. skills](7.%20skills.md)). |
| **Multi-provider models** | More than **one** in-tree provider (e.g. Anthropic) is optional; the **interface** should still allow adding providers without breaking `agent`. Each provider must still support the **streaming** contract used by `agent`. |
| **Teams / workflows** | Single `Runner` only ([8. architecture § Future extensions](8.%20architecture.md#future-extensions-non-foundation)). |
| **Production ingress** | No mandated HTTP/gRPC **server** in v1; a CLI or library-only integration is enough. Auth, rate limits, and multi-tenant policy are **host app** concerns. |
| **Postgres / Redis stores** | Optional later; **memory** (+ optional **SQLite**) suffices for MVP ([4. databases](4.%20databases.md)). |
| **Advanced safety** | Content filters, human-in-the-loop approvals, and policy engines are **optional hooks** at most—not blocking MVP. |

---

## Package and dependency expectations

Implementation should follow the **package map** in [9. code-structure §2](9.%20code-structure.md#2-package-map-ownership-and-imports):

- **`internal/session`** — Domain types and `Store` **interfaces**; avoid importing `agent`, `model`, `tool`, or concrete `store/*`.
- **`internal/agent`** — Orchestrator + tool loop; depends on `session`, `prompt`, `model`, `tool`, and optionally **`skill`** only when a stub exists.
- **`internal/prompt`** — Pure assembly; may import `session` only.
- **`internal/model`** — HTTP adapter(s) with **streaming** request/response handling; may use `session` types; must not import `tool` or `agent`.
- **`internal/tool`** — Registry and execution; may import `session`; must not import `agent`.
- **`internal/store/...`** — Implements `session.Store`; may import `session` and drivers; must not import `agent`.

If MVP code **must** diverge from this graph, update **docs/8** and **docs/9** in the same change set and note the exception in the PR summary ([implement-centrai-go](../.cursor/skills/implement-centrai-go/SKILL.md)).

---

## Configuration

| Layer | MVP expectation |
|-------|-----------------|
| **`cmd/*`** | Env/flags only; validate and map into structs. |
| **`internal/*`** | No hidden `os.Getenv` in libraries; config passed in via constructors ([9. code-structure §8](9.%20code-structure.md#8-configuration-layers)). |

---

## Streaming semantics (MVP)

- **Inbound**: The model client consumes the provider stream, **forwards** chunk/delta events to the caller (channel, callback, or iterator—implementation choice), and **buffers or parses** until it can produce a single **assistant step**: final text segment(s) and/or **tool calls** as defined by the provider (some APIs emit tool calls only after the stream completes).
- **Outbound**: Hosts (CLI, library) should be able to **observe streaming progress** for UX (e.g. incremental text); the **tool loop** still runs on **completed** assistant steps, not on partial tool-call fragments, unless the chosen API defines otherwise.
- **Tests**: Use **`httptest`** with **synthetic SSE or chunked bodies** so streaming paths are covered without a live provider.

---

## Observability (minimal)

- Emit **structured lifecycle events** where cheap (e.g. run start/end, model stream start/end, tool start/end)—even if initially logged via an injected `slog.Logger` or narrow interface ([8. architecture § Observability](8.%20architecture.md#observability)).
- Do **not** log raw secrets or full prompts in production defaults.

---

## Acceptance criteria (checklist)

Use this list to decide when MVP is **complete** for a release tag.

- [x] `go.mod` present; module path documented in README or `agent` package doc.
- [x] `Runner` (or equivalent) exposes **`Run(ctx, *RunInput) (*RunOutput, error)`** (or same semantics) with **max iterations** and **context** respected; **every model invocation uses the streaming API path**.
- [x] At least **one** model implementation drives a **streaming** chat request, emits incremental events, and produces final assistant text and/or tool calls for the loop.
- [x] **Tool loop** validates names/arguments, executes handlers, appends tool messages, stops on completion or limit.
- [x] **`session.Store`** persists append-only message history for a session id (in-memory minimum).
- [x] **Tests** cover core loop and model HTTP layer without live network (httptest + fakes).
- [x] **Docs**: README **Status** reflects that code exists; this file updated if scope changes.

---

## Related documents

| Topic | Document |
|-------|----------|
| Architecture | [8. architecture](8.%20architecture.md) |
| Code structure (Go) | [9. code-structure](9.%20code-structure.md) |
| Agents | [1. agents](1.%20agents.md) |
| Models | [2. models](2.%20models.md) |
| Tools | [3. tools](3.%20tools.md) |
| Sessions | [5. session-management](5.%20session-management.md) |

---

*See also: [README](../README.md) · [8. architecture](8.%20architecture.md) · [9. code-structure](9.%20code-structure.md)*
