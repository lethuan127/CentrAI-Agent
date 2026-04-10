# Contributing to CentrAI Agent

Thank you for your interest in this project. This document describes how to work with the codebase and propose changes.

## Before you start

- Read [README.md](README.md) for scope and documentation links.
- For architecture and package boundaries, see [docs/8. architecture.md](docs/8.%20architecture.md) and [docs/9. code-structure.md](docs/9.%20code-structure.md).
- Agent and Cursor conventions are summarized in [AGENTS.md](AGENTS.md).

## How to contribute

1. **Open an issue** (or comment on an existing one) for non-trivial work so maintainers can agree on direction.
2. **Fork** the repository and create a **topic branch** from the default branch.
3. **Make focused changes** — one logical change per pull request when possible.
4. **Run checks locally** (see below) before opening a PR.
5. **Open a pull request** with a clear description of the problem and the solution.

## Local development

```bash
go test ./... -race -count=1
go vet ./...
```

Ensure Go code is formatted (`gofmt` / `go fmt`). CI runs the same checks (format, vet, tests with `-race`) in [`.github/workflows/ci.yml`](.github/workflows/ci.yml).

## API and documentation

- If you change user-visible HTTP behavior, update [api/openapi.yaml](api/openapi.yaml) when applicable.
- If you add or rename top-level design docs under `docs/`, update the documentation table in [README.md](README.md) per project conventions.

## Code style

- Match existing naming, layering, and dependency direction described in `docs/8` and `docs/9`.
- Prefer constructor injection over globals for new code.
- The primary agent execution path uses **streaming** model I/O; see [docs/V2.md](docs/V2.md).

## Security

See [SECURITY.md](SECURITY.md) for how to report vulnerabilities responsibly.

## Code of conduct

Participants are expected to follow [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).
