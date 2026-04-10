# CentrAI Agent

Open-source **agent runtime**: a stateful control loop around a stateless language model—plans steps, calls tools, persists sessions, and stops when the task is done. The design follows the same core ideas as [Agno’s agent model](https://docs.agno.com/agents/overview) (tools, instructions, memory, storage), but this project is implemented in **Go** for performance, predictable latency, and efficient concurrency in production services.

## Why Go

- **Throughput and memory** — Tighter footprint than typical interpreted stacks for hot paths (routing, session handling, tool dispatch).
- **Concurrency** — First-class goroutines and channels fit streaming runs, parallel tool execution, and MCP clients.
- **Deployment** — Single static binaries simplify containers, edge, and sidecar deployments.

The LLM still lives behind HTTP APIs; Go owns orchestration, I/O, and state.

## Documentation

| Topic | Doc |
|-------|-----|
| Overview | [docs/1. agents.md](docs/1.%20agents.md) |
| MVP (first version scope) | [docs/MVP.md](docs/MVP.md) |
| Version 2 (standards, integration, persistence) | [docs/V2.md](docs/V2.md) |
| Architecture (Go) | [docs/8. architecture.md](docs/8.%20architecture.md) |
| Code structure (Go) | [docs/9. code-structure.md](docs/9.%20code-structure.md) |
| Models & providers | [docs/2. models.md](docs/2.%20models.md) |
| Tools | [docs/3. tools.md](docs/3.%20tools.md) |
| Databases / persistence | [docs/4. databases.md](docs/4.%20databases.md) |
| Session management | [docs/5. session-management.md](docs/5.%20session-management.md) |
| MCP | [docs/6. mcps.md](docs/6.%20mcps.md) |
| Skills | [docs/7. skills.md](docs/7.%20skills.md) |
| Agent definitions (CLI / library) | [.centrai/agents/README.md](.centrai/agents/README.md) |
| AI assistants / Cursor | [AGENTS.md](AGENTS.md) |

Start with **agents**, then follow the links in each file.

## Status

Go module **`github.com/lethuan127/centrai-agent`** is in-repo: streaming model client (OpenAI-compatible), tool registry, in-memory session store, run orchestrator, and `cmd/centrai` CLI. See [docs/MVP.md](docs/MVP.md) for scope and acceptance criteria.

## License

To be determined (TBD).
