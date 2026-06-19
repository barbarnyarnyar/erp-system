---
name: contextkit
description: Use ContextKit to save or resume project memory for any Git repository. Trigger when the user asks to use ContextKit, save project memory, compress context, create a handoff note, prepare next agent handoff, resume with ContextKit, load project memory, read project context, or start from the last handoff.
version: 1.0.0
---

# ContextKit

ContextKit is the project memory save and resume skill.

Use it to preserve continuity across token limits, interrupted sessions, tool switching, multi-day gaps, and agent handoffs. The skill source lives in `metis-lab`, but every project keeps its own live memory files inside that project's repo.

ContextKit is not a feature-building workflow. During save or resume, do not edit production code, add dependencies, refactor, redesign, run destructive Git commands, delete branches, rewrite history, or expose secrets unless the user explicitly asks for that work separately.

## Choose The Workflow

Use **Save / Compress** when the user says:

- use ContextKit
- save project memory
- compress context
- create handoff note
- prepare next agent handoff
- update project context

Use **Resume / Retrieve** when the user says:

- resume with ContextKit
- load project memory
- read project context
- catch me up
- what is the latest project state?
- start from the last handoff

If the request is ambiguous, prefer Resume / Retrieve first. It is safer to read and summarize before writing memory files.

## Repo Memory Files

ContextKit maintains these files in the active project repo:

```text
AGENTS.md
docs/
  PROJECT_CONTEXT.md
  DECISION_LOG.md
  SESSION_NOTES/
    YYYY-MM-DD-session.md
```

Do not create one central memory file for all projects. Each repo owns its own memory so future agents can resume after cloning or pulling that repo.

## Save / Compress Workflow

1. Find the active Git repo root with `git rev-parse --show-toplevel`.
2. Inspect:
   - `git status --short --branch`
   - current branch
   - `git log --oneline -10`
   - changed files from `git status` and targeted diffs as needed
3. Read existing files if present:
   - `AGENTS.md`
   - `docs/PROJECT_CONTEXT.md`
   - `docs/DECISION_LOG.md`
   - latest file in `docs/SESSION_NOTES/`
4. Create missing memory directories or files only as needed.
5. Update `docs/PROJECT_CONTEXT.md` with current high-level truth only.
6. Update `docs/DECISION_LOG.md` only when a real decision was made.
7. Create or update today's `docs/SESSION_NOTES/YYYY-MM-DD-session.md`.
8. Add the ContextKit memory protocol to `AGENTS.md` if missing.
9. Reply with files updated, current phase, locked decisions, known risks, next task, and suggested next prompt.

### PROJECT_CONTEXT.md Shape

Keep this file short: current state, not a diary. Target roughly 80 lines or less.

```md
# PROJECT_NAME Project Context

## Current phase
...

## Current branch
...

## Current status
...

## Current priorities
1. ...
2. ...
3. ...

## Do not reopen
- ...

## Known risks
- ...

## Most relevant files
- ...

## Next best task
...
```

### DECISION_LOG.md Shape

Only log decisions that prevent future rework or explain a meaningful project direction.

```md
# PROJECT_NAME Decision Log

## YYYY-MM-DD - Decision title

Decision: ...
Reason: ...
Status: Locked / Accepted / Deferred / Rejected
Reopen only if: ...
```

Do not log tiny implementation details. Do not mark anything as locked unless the user, Git history, or repo content clearly supports it.

### SESSION_NOTES Shape

Use the current machine date for the filename.

```md
# Session Note - YYYY-MM-DD

## Branch
...

## Session goal
...

## Changed
- ...

## Files touched
- ...

## Decisions made
- ...

## Known issues
- ...

## Next session should start with
1. ...
2. ...
3. ...

## Suggested next prompt
...
```

Session notes are handoff notes. Keep them practical, compressed, and useful to an agent without chat history.

## Resume / Retrieve Workflow

1. Find the active Git repo root with `git rev-parse --show-toplevel`.
2. Read:
   - `AGENTS.md`
   - `docs/PROJECT_CONTEXT.md`
   - `docs/DECISION_LOG.md`
   - latest file in `docs/SESSION_NOTES/`
3. Inspect:
   - `git status --short --branch`
   - `git log --oneline -10`
4. Summarize:
   - current phase
   - active branch
   - latest changed files
   - locked decisions
   - known risks or blockers
   - most relevant files
   - safest next task
5. Do not edit files unless the user asks to save or compress context.

If memory files are missing, say so and summarize from Git state. Offer to run Save / Compress to initialize ContextKit memory.

## AGENTS.md Protocol

When initializing or updating `AGENTS.md`, add this section if missing:

```md
## ContextKit project memory protocol

This repo uses ContextKit memory files to preserve project continuity across AI sessions.

Before starting meaningful work, read:

- `docs/PROJECT_CONTEXT.md`
- `docs/DECISION_LOG.md`
- latest file in `docs/SESSION_NOTES/`

For resume requests, read the memory files, inspect Git state, and report the current phase, branch, locked decisions, known risks, and safest next task before editing.

For save/compression requests, inspect Git state, review changed files as needed, update the memory docs, and do not change production code unless explicitly asked.
```

## Evidence Rules

- Treat Git state and repo files as the source of truth.
- Do not invent progress.
- Do not mark work done unless Git, files, or the user confirms it.
- Write `needs verification` when uncertain.
- Remove stale memory when the current repo state contradicts it.
- Keep memory factual and compact.
- Never include secrets, tokens, credentials, private keys, or sensitive customer data.

## Reference

For the full product spec, read `references/contextkit_project_memory_skill_prd.md` only when creating, changing, or reviewing the ContextKit skill itself.
