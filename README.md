# CentrAI Agent

Open-source **agent runtime**: a stateful control loop around a stateless language model—plans steps, calls tools, persists sessions, and stops when the task is done. It centers on **tools**, **instructions**, **memory**, and **storage** behind small interfaces; the implementation is **Go** for performance, predictable latency, and efficient concurrency in production services.

## Why Go

- **Throughput and memory** — Tighter footprint than typical interpreted stacks for hot paths (routing, session handling, tool dispatch).
- **Concurrency** — First-class goroutines and channels fit streaming runs, parallel tool execution, and MCP clients.
- **Deployment** — Single static binaries simplify containers, edge, and sidecar deployments.

The LLM still lives behind HTTP APIs; Go owns orchestration, I/O, and state.

## Documentation

Design and behavior are documented under **`docs/`** (see [docs/8. architecture.md](docs/8.%20architecture.md) and [docs/9. code-structure.md](docs/9.%20code-structure.md) for the big picture). Use the table below to jump to a topic.

| Topic | Doc |
|-------|-----|
| **Contributing** | [CONTRIBUTING.md](CONTRIBUTING.md) |
| **Plan** (roadmap + what is implemented today) | [docs/plan.md](docs/plan.md) |
| Overview | [docs/1. agents.md](docs/1.%20agents.md) |
| Architecture (Go) | [docs/8. architecture.md](docs/8.%20architecture.md) |
| Code structure (Go) | [docs/9. code-structure.md](docs/9.%20code-structure.md) |
| Models & providers | [docs/2. models.md](docs/2.%20models.md) |
| Tools | [docs/3. tools.md](docs/3.%20tools.md) |
| Databases / persistence | [docs/4. databases.md](docs/4.%20databases.md) |
| Session management | [docs/5. session-management.md](docs/5.%20session-management.md) |
| MCP | [docs/6. mcps.md](docs/6.%20mcps.md) |
| Skills | [docs/7. skills.md](docs/7.%20skills.md) |
| HTTP API (OpenAPI) | [api/openapi.yaml](api/openapi.yaml) |
| Agent definitions (CLI / library) | [.centrai/agents/README.md](.centrai/agents/README.md) |
| AI assistants / Cursor | [AGENTS.md](AGENTS.md) |

**Suggested reading order for new contributors:** [CONTRIBUTING.md](CONTRIBUTING.md) → [docs/1. agents.md](docs/1.%20agents.md) (foundation series 1–7 as needed) → [docs/8. architecture.md](docs/8.%20architecture.md) and [docs/9. code-structure.md](docs/9.%20code-structure.md) → [docs/plan.md](docs/plan.md) for the code map and roadmap.

## Status

Go module **`github.com/lethuan127/centrai-agent`** is in-repo: streaming model clients (OpenAI-compatible chat and Responses API profiles), tool registry, session stores (memory and SQLite), run orchestrator, MCP and skill packages, optional HTTP ingress (`internal/httpserver`, `api/openapi.yaml`), and `cmd/centrai` CLI. See [docs/plan.md](docs/plan.md) for a **current implementation snapshot** and remaining roadmap items.

## License

[MIT](LICENSE). Community: [CONTRIBUTING.md](CONTRIBUTING.md), [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md), [SECURITY.md](SECURITY.md).
