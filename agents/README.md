# Agent definitions (YAML)

Declarative agent files (similar in spirit to **Cursor** agent rules, **Claude Code** `CLAUDE.md` / subagent specs, and **Codex**-style config): one YAML file per agent.

## Layout

| Field | Required | Description |
|-------|----------|-------------|
| `version` | yes | Must be `1`. |
| `kind` | no | If set, must be `Agent`. |
| `name` | no | Short id for logging. |
| `description` | no | Prepended before `instructions` in the system prompt when non-empty. |
| `instructions` | yes | System prompt (role, behavior, formatting). |
| `model` | no | Overrides the chat model (CLI/env default applies if empty). |
| `tools` | no | Tool bundles: `demo` registers `echo` and `add` (see `cmd/centrai`). |
| `max_steps` | no | Max model rounds per user turn (> 0). |
| `metadata` | no | Arbitrary string map for hosts or tooling. |

## Example

See [example.yaml](example.yaml).

## CLI

```bash
go run ./cmd/centrai -agent agents/example.yaml -repl
```

Use `-message "..."` instead of `-repl` for a single turn.
