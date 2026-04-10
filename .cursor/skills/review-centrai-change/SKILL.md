---
name: review-centrai-change
description: Reviews proposed diffs or PRs for CentrAI Agent against documented architecture, package boundaries, and documentation consistency. Use when the user asks for a code review, PR review, or sanity check before merge in this repository.
---

# Review CentrAI changes

## Check

1. **Layers**: Does new Go code respect the dependency direction in [docs/9. code-structure.md](../../../docs/9.%20code-structure.md) (e.g. `session` not importing `agent`)?
2. **Placement**: Does each package still have a single clear job? No new `util` / `common` packages without strong justification.
3. **APIs**: Blocking calls take `context.Context` first; errors wrapped; config from `cmd` / loaders, not hidden env reads in `internal`.
4. **Docs**: If behavior or layout changed, are [docs/8](../../../docs/8.%20architecture.md) / [docs/9](../../../docs/9.%20code-structure.md), [docs/plan](../../../docs/plan.md) (roadmap vs tree), and [README](../../../README.md) table updated?

## Output

Give **must-fix** vs **nice-to-have**; cite the relevant doc section or package rule when flagging an issue.
