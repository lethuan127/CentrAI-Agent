---
name: implement-centrai-go
description: Implements or refactors CentrAI Agent Go code using the prescribed internal layout (agent, session, prompt, model, tool, mcp, store, skill) and dependency rules. Use when adding Go packages, the run/tool loop, model clients, session storage, MCP client, skill loader, or cmd entrypoints in this repository.
---

# Implement CentrAI (Go)

## Before coding

1. Read [docs/9. code-structure.md](../../../docs/9.%20code-structure.md) for package ownership and the import graph.
2. Read [docs/8. architecture.md](../../../docs/8.%20architecture.md) for layers and runtime sequence.
3. Do not modify `temp/agno/` for product work — reference only.

## Implementation checklist

- [ ] New code lives under the correct `internal/` package; `cmd/` only wires config and dependencies.
- [ ] Dependencies flow downward: `session` types/interfaces stay free of `agent` / `model` / `tool` imports as specified in the doc table.
- [ ] Blocking functions take `context.Context` first; errors wrapped; stable sentinel errors where comparisons matter.
- [ ] Stores implement `session.Store` in `internal/store/*`; no upward import to `agent`.
- [ ] Tests colocated; HTTP with `httptest`; integration behind `integration` build tag if needed.

## Escalation

If the design requires breaking the documented dependency graph, update `docs/8` and `docs/9` in the same change set and say so in the summary.
