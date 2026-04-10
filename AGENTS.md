# Agent notes (CentrAI Agent)

- **Design source of truth**: [docs/](docs/) — start with [docs/8. architecture.md](docs/8.%20architecture.md) and [docs/9. code-structure.md](docs/9.%20code-structure.md).
- **Do not treat** `temp/agno/` as product code to change; reference only.
- **Cursor index**: [.cursorignore](.cursorignore) excludes `temp/` from indexing.

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
