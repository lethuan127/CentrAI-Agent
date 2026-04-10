---
name: maintain-centrai-docs
description: Edits or extends CentrAI Agent Markdown documentation under docs/ while keeping cross-links, numbering, and architecture alignment consistent. Use when updating design docs, adding a new doc file, reconciling docs with planned Go layout, or reviewing doc consistency.
---

# Maintain CentrAI docs

## Steps

1. Open the doc(s) that define the same topic (e.g. architecture vs code structure) and align terminology (orchestrator, run, tool loop, session store).
2. Preserve link style: relative paths, spaces encoded as `%20` in filenames where the repo already does so.
3. Update the README doc table in [README.md](../../../README.md) if you add a new top-level doc or rename one.
4. If docs describe APIs or directories that do not exist in git yet, state that clearly (status/roadmap) so readers are not misled.
5. If you add, rename, or remove **Cursor rules or skills**, update the inventory in [AGENTS.md](../../../AGENTS.md).

## Checklist

- [ ] No contradiction between `docs/8` (layers) and `docs/9` (packages/imports).
- [ ] Related documents sections and next/previous navigation updated when files are added or reordered.
- [ ] Mermaid or diagrams remain valid if layers or names change.
- [ ] If `.cursor/rules/` or `.cursor/skills/` changed, [AGENTS.md](../../../AGENTS.md) inventory is updated.
