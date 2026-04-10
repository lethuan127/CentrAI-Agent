# Contributing to CentrAI Agent

Thank you for your interest in this project. This document is the **entry point for contributors**: how to build, test, read the design, and open a solid pull request.

## Quick start

1. **Install Go** — Use the version in [`go.mod`](go.mod) (currently **1.25+**).
2. **Clone** this repository and work from the repo root.
3. **Verify** your setup:

   ```bash
   go test ./... -race -count=1
   go vet ./...
   ```

4. **Read the design** before large changes — see [What to read first](#what-to-read-first) below.

## What to read first

| Order | Document | Why |
|-------|----------|-----|
| 1 | [README.md](README.md) | Scope, documentation index, status |
| 2 | [docs/8. architecture.md](docs/8.%20architecture.md) | Layers, streaming, boundaries |
| 3 | [docs/9. code-structure.md](docs/9.%20code-structure.md) | Packages, imports, where code belongs |
| 4 | [docs/plan.md](docs/plan.md) | What exists in-tree today and what is still planned |

Use the foundation series ([docs/1. agents.md](docs/1.%20agents.md) through [docs/7. skills.md](docs/7.%20skills.md)) when you touch that area.

**Cursor / AI assistants:** conventions for this repo are summarized in [AGENTS.md](AGENTS.md).

## Where things live

| Location | Contents |
|----------|----------|
| [`docs/`](docs/) | Design docs (source of truth for architecture and behavior) |
| [`internal/`](internal/) | Runtime implementation (not importable by other modules) |
| [`cmd/centrai/`](cmd/centrai/) | CLI entrypoint and wiring only |
| [`api/openapi.yaml`](api/openapi.yaml) | HTTP API contract (when REST routes change) |
| [`.centrai/`](.centrai/) | Example agent defs, local skills, MCP notes (not core library code) |

## How to contribute

1. **Open an issue** (or comment on an existing one) for non-trivial work so maintainers can agree on direction.
2. **Fork** the repository and create a **topic branch** from the default branch.
3. **Make focused changes** — one logical change per pull request when possible.
4. **Run checks locally** (see [Quick start](#quick-start)) before opening a PR.
5. **Open a pull request** using the [pull request template](.github/pull_request_template.md): describe the problem, the solution, and any doc or API updates.

## API and documentation

- **HTTP behavior** — If user-visible routes or payloads change, update [`api/openapi.yaml`](api/openapi.yaml) in the same change when applicable.
- **Design docs** — If you add or rename a **top-level** file under `docs/` (e.g. new `docs/Foo.md`), update the documentation table in [README.md](README.md).
- **Cursor rules or skills** — If you add or rename files under `.cursor/rules/` or `.cursor/skills/`, update the inventory in [AGENTS.md](AGENTS.md).

## Code style

- Match existing naming, layering, and **dependency direction** in `docs/8` and `docs/9`.
- Prefer **constructor injection** over globals for new code.
- The primary agent execution path uses **streaming** model I/O; see [docs/plan.md](docs/plan.md#streaming-rule-agent-run).

## Security

See [SECURITY.md](SECURITY.md) for how to report vulnerabilities responsibly.

## Code of conduct

Participants are expected to follow [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).
