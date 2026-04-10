# Plan — roadmap and repository state

This document is the **living plan**: what is **implemented today** in this repository, what **streaming and API rules** we keep, and what remains for **deeper integration and production hardening**. It aligns with [8. architecture](8.%20architecture.md) and [9. code-structure](9.%20code-structure.md); foundation topics remain in [1. agents](1.%20agents.md)–[7. skills](7.%20skills.md).

**Relationship to the core loop:** The **`Runner`** contract (streaming model, tool loop, `session.Store`) is the stable center. Work below **extends** hosts and operators without redefining that loop unless design docs change in the same PR.

---

## Public Go module (`package centrai`)

The root package **[`centrai.go`](../centrai.go)** is the supported **library surface** for importers: it re-exports types and constructors so hosts can wire their own binaries.

| Capability | API / package | Notes |
|------------|---------------|--------|
| **Run orchestration** | [`internal/agent`](../internal/agent/) via `centrai.NewRunner` | Tool loop, `RunInput` / `RunOutput`, optional `OnModelChunk`, `TraceID`. |
| **OpenAI-compatible streaming** | [`internal/model/openai`](../internal/model/openai/) via `centrai.NewOpenAIClient` | Chat Completions–style streaming HTTP. |
| **OpenAI Responses API streaming** | [`internal/model/responsesapi`](../internal/model/responsesapi/) via `centrai.NewResponsesAPIClient` | Second in-tree `model.Client` implementation. |
| **Session stores** | [`internal/store/memory`](../internal/store/memory/), [`internal/store/sqlite`](../internal/store/sqlite/) via `centrai.NewMemoryStore`, `centrai.NewSQLiteStore` | SQLite is file-backed; callers own lifecycle / `Close`. |
| **Tools** | [`internal/tool`](../internal/tool/) | Registry, JSON Schema, middleware (e.g. `BlockToolNames` exported on `centrai`). |
| **MCP → registry** | [`internal/mcp`](../internal/mcp/) via `centrai.RegisterMCPTools` | Requires a live `mcp.ClientSession` from the MCP Go SDK; maps remote tools into `tool.Registry`. |
| **Skills (filesystem)** | [`internal/skill`](../internal/skill/) | `skill.Loader` resolves paths / search dirs to instruction text—**host must call** and merge into system/instructions. |
| **Agent definitions** | [`internal/agentdef`](../internal/agentdef/) via `centrai.LoadAgentDefinition` | YAML or Markdown + YAML front matter. |
| **Prompt helpers** | [`internal/prompt`](../internal/prompt/) | e.g. `MergeSystemWithSession`, `FormatSessionForSystem`. |

Shared streaming types live in [`internal/model/client.go`](../internal/model/client.go).

---

## Reference CLI (`cmd/centrai`)

The **`cmd/centrai`** binary is a **minimal demo / dev entrypoint**, not an exhaustive showcase of every package. Treat it as **one** wiring of `Runner`; production apps typically import `centrai` and compose their own store and model.

| Topic | Actual behavior |
|-------|-----------------|
| **Model** | **Only** [`internal/model/openai`](../internal/model/openai/) — OpenAI-compatible HTTP (`-base-url`, `-model`, `OPENAI_API_KEY`). No flag to select Responses API; use **`centrai.NewResponsesAPIClient`** in your own `main`. |
| **Store** | **Only** in-memory [`memory.New()`](../internal/store/memory/memory.go). No `-sqlite` / DB path flag; use **`centrai.NewSQLiteStore`** (or your store) in custom code. |
| **Modes** | One-shot `-message`, `-repl`, or **HTTP** `-http ADDR` (see below). Requires `-message`, `-repl`, or `-http` (mutually exclusive for the non-HTTP paths). |
| **Agent file** | Optional `-agent` → [`agentdef.LoadFile`](../internal/agentdef/agentdef.go); adjusts system text, `maxTurns`, demo tools via definition. |
| **MCP / skills in agent YAML** | `mcpServers` and `skills` in an agent definition are reflected in **metadata lines** appended to the system prompt (`LLMMetaAppendix`) — they are **not** auto-connected to MCP transports or `skill.Loader` in this CLI. Integrators call `RegisterMCPTools` / `Loader` themselves. |
| **Env** | `OPENAI_API_KEY` (required unless `CENTRAI_SKIP_API=1`), optional `OPENAI_BASE_URL`, `OPENAI_MODEL` (defaults also set in [`config.go`](../cmd/centrai/config.go)). |

Flags and env are defined in [`cmd/centrai/config.go`](../cmd/centrai/config.go); behavior in [`cmd/centrai/main.go`](../cmd/centrai/main.go).

---

## Optional HTTP ingress

[`internal/httpserver`](../internal/httpserver/) registers:

- **`POST /v1/run`** — JSON body (`session_id`, `message`, optional `system`, `trace_id`, `max_steps`, `model`); **SSE** stream of `chunk` / `done` / `error` events; uses `Runner` with `OnModelChunk` for deltas.
- **`GET`/`HEAD /healthz`** — liveness.

Wired from `cmd/centrai` when `-http` is set; contract in [`api/openapi.yaml`](../api/openapi.yaml). **`trace_id`** may be filled from JSON or from the `traceparent` header (see handler).

---

## Streaming rule (agent run)

Every **Run** must drive the model through a **streaming** API path only—token/chunk (or equivalent delta) events through to a **completed** assistant step for the tool loop. There is **no** supported first-class **non-streaming model completion** path on `model.Client`, ingress, or the orchestrator for normal agent execution. HTTP **error responses** (4xx/5xx) remain synchronous as usual; only **successful model output** is stream-shaped.

---

## Roadmap themes (remaining work)

| Theme | Target |
|-------|--------|
| **CLI parity (optional)** | Flags or subcommands to exercise SQLite, Responses API, or MCP from `cmd/centrai` if we want the binary to mirror the library—today the **library** is the full surface. |
| **Standard external surfaces** | Explicit **stability** tiers for import paths; README versioning; **OpenAPI** kept in lockstep with [`api/openapi.yaml`](../api/openapi.yaml). |
| **MCP** | End-to-end example: connect transport → `RegisterMCPTools`; operational **timeouts / auth / naming** docs in [6. mcps](6.%20mcps.md). |
| **Skills** | Example host wiring: `skill.Loader` + agentdef paths; optional **registry** resolution ([7. skills](7.%20skills.md)). |
| **Persistence** | **PostgreSQL** (or equivalent) adapter and **migrations** for SQL backends ([4. databases](4.%20databases.md)). |
| **Ingress** | gRPC or additional HTTP route families; AI UI **compatibility** notes ([HTTP, gRPC, and AI UI](#http-grpc-and-ai-ui-stacks) below). |
| **Observability** | Correlation ids end-to-end beyond current `TraceID` / SSE; stable lifecycle event catalog ([8. architecture § Observability](8.%20architecture.md#observability)). |
| **Safety** | Document default-redacting middleware and policy patterns ([3. tools](3.%20tools.md)). |

---

## Out of scope (near term)

| Area | Rationale |
|------|-----------|
| **Teams / multi-agent workflows** | Compose multiple `Runner` instances later ([8. architecture § Future extensions](8.%20architecture.md#future-extensions-non-foundation)). |
| **Built-in vector RAG / embeddings** | Adapter / product layers ([4. databases](4.%20databases.md)). |
| **Full SaaS control plane** | Host concern. |
| **Non-streaming model completion for Run** | Callers may **buffer** a stream; no non-stream `model.Client` mode for agent execution ([Streaming rule](#streaming-rule-agent-run)). |

---

## HTTP, gRPC, and AI UI stacks

**Today:** REST + SSE for `POST /v1/run` per OpenAPI.

**Plan:** Optional gRPC + **compatibility matrix** for AI UI transports when we commit to a path. **Go-first**; TS UI stays in host repos.

---

## Documentation expectations

- Foundation docs should describe **landed** library behavior; this file tracks **CLI vs library** and **gaps**.
- **Layers / packages:** [8. architecture](8.%20architecture.md), [9. code-structure](9.%20code-structure.md).

---

## Integration checklist

Use this list for a broader **integration / productization** milestone.

**Library (in-tree)**

- [x] Streaming **`Runner`**; max steps; `context` cancellation; streaming-only `model.Client` path.
- [x] Two **`model.Client`** implementations: `openai`, `responsesapi`.
- [x] **`session.Store`**: memory + SQLite implementations (callers choose).
- [x] **Tool** registry + middleware.
- [x] **MCP** package + `RegisterMCPTools` (host supplies MCP session).
- [x] **`skill.Loader`** (host merges text into prompts).
- [x] **HTTP** handler + **OpenAPI** for `/v1/run` and `/healthz`.

**Reference CLI (`cmd/centrai`) — intentionally narrow**

- [x] OpenAI-compatible client + memory store + `-agent` / demo tools / `-http`.
- [ ] Does **not** yet switch model backend, SQLite, MCP, or skill file loading—by design; use **`package centrai`** or extend `cmd`.

**Still open (roadmap)**

- [ ] PostgreSQL (or HA) store + migrations.
- [ ] Broader observability catalog and docs.
- [ ] Public API stability / semver story in README.
- [ ] AI SDK / UI adapter recipe or matrix.
- [ ] CHANGELOG / release discipline.

---

## Related documents

| Topic | Document |
|-------|----------|
| Architecture | [8. architecture](8.%20architecture.md) |
| Code structure (Go) | [9. code-structure](9.%20code-structure.md) |
| Agents | [1. agents](1.%20agents.md) |
| Models | [2. models](2.%20models.md) |
| Tools | [3. tools](3.%20tools.md) |
| Databases / persistence | [4. databases](4.%20databases.md) |
| Sessions | [5. session-management](5.%20session-management.md) |
| MCP | [6. mcps](6.%20mcps.md) |
| Skills | [7. skills](7.%20skills.md) |

**Implementation reminders:** The **streaming rule** is mirrored in [.cursor/rules/go-centrai.mdc](../.cursor/rules/go-centrai.mdc) and [.cursor/rules/centrai-core.mdc](../.cursor/rules/centrai-core.mdc).

---

*See also: [README](../README.md) · [8. architecture](8.%20architecture.md) · [9. code-structure](9.%20code-structure.md)*
