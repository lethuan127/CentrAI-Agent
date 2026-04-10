---
name: review-centrai-change
description: Pre-merge review against docs/8, docs/9, and MVP scope for CentrAI Agent.
---

# Review CentrAI change

1. Read [docs/8. architecture.md](../../docs/8.%20architecture.md) and [docs/9. code-structure.md](../../docs/9.%20code-structure.md) for package boundaries.
2. Confirm `internal/` import direction matches the package map (session does not import agent).
3. For user-visible behavior, check [docs/MVP.md](../../docs/MVP.md) and [docs/V2.md](../../docs/V2.md) scope.
4. Prefer constructor-injected config; no hidden `os.Getenv` in libraries (`cmd` only).
