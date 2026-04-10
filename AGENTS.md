# Agent notes (CentrAI Agent)

- **Contributing (humans)**: [CONTRIBUTING.md](CONTRIBUTING.md) — build, test, reading order, repo layout.
- **Roadmap + current code map**: [docs/plan.md](docs/plan.md).
- **Design source of truth**: [docs/](docs/) — start with [docs/8. architecture.md](docs/8.%20architecture.md) and [docs/9. code-structure.md](docs/9.%20code-structure.md).
- **Cursor index**: [.cursorignore](.cursorignore) excludes `temp/` from indexing (optional local files only; not product code).

## Project rules ([.cursor/rules/](.cursor/rules/))

| File | When it applies |
|------|-----------------|
| `centrai-core.mdc` | Always (`alwaysApply`) — identity, canonical docs, constraints |
| `go-centrai.mdc` | When `**/*.go` files are in context |
| `centrai-documentation.mdc` | When `docs/**/*.md` is in context |
| `readme-centrai.mdc` | When root `README.md` is in context |

## Project skills ([.cursor/skills/](.cursor/skills/))

| Skill | Use for |
|-------|---------|
| [implement-centrai-go](.cursor/skills/implement-centrai-go/SKILL.md) | Adding or changing Go runtime code |
| [maintain-centrai-docs](.cursor/skills/maintain-centrai-docs/SKILL.md) | Editing or extending `docs/` |
| [review-centrai-change](.cursor/skills/review-centrai-change/SKILL.md) | PR / pre-merge review against architecture |
