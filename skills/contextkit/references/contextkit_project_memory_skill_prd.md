# PRD - ContextKit Agent Skill: Project Memory Save + Resume

## 1. Product Summary

Build a lightweight reusable agent skill that allows Codex or another AI coding agent to save and resume project memory for any Git repository.

ContextKit preserves continuity across token limits, interrupted sessions, tool switching, multi-day gaps, and handoffs between AI agents. It does this by maintaining short, structured Markdown memory files inside the active project repo.

The skill source and version history live in `metis-lab`. Live project memory always lives in the project repo being worked on.

## 2. Problem Statement

AI-assisted project work often spans multiple sessions, tools, branches, and days. When a session ends because of token limits, tool limits, fatigue, or context loss, the next session often has to reconstruct:

- what was already decided
- what should not be reopened
- what branch is active
- what changed recently
- what remains broken
- what the safest next task is
- which files matter most

Without repo-native memory, agents waste time rebuilding state from chat history, Git graph, or human memory. The project repo should carry enough context for a future agent to resume after cloning, pulling, or opening the project.

## 3. Goals

- Provide a reusable project memory skill for any Git repo under the AntiGravity workspace or global Codex environment.
- Support both context saving and context retrieval.
- Keep each project self-contained by storing live memory inside that project's repo.
- Keep memory short, factual, and useful for the next AI session.
- Avoid product-code changes during memory save/resume workflows unless explicitly requested.

## 4. Non-Goals

ContextKit must not:

- implement product features
- redesign pages, flows, or user interfaces
- refactor production code
- add dependencies
- run destructive Git commands
- delete branches
- rewrite commit history
- invent progress that cannot be verified
- mark incomplete work as done
- replace Git commits, issues, project boards, release tags, or changelogs
- maintain one central memory file for all projects

## 5. Target Users

Primary user: Aung Shwe Oo, working across multiple AI-assisted projects in the AntiGravity workspace.

Secondary user: Any AI coding agent that needs to save or resume a project after context loss.

## 6. Architecture

ContextKit has two layers.

### Source layer

The reusable skill source lives in `metis-lab`:

```text
metis-lab/skills/contextkit/
  SKILL.md
  agents/openai.yaml
  references/contextkit_project_memory_skill_prd.md
```

This is where future versions of the skill are improved.

### Runtime memory layer

Each active project repo keeps its own live memory:

```text
PROJECT_REPO/
  AGENTS.md
  docs/
    PROJECT_CONTEXT.md
    DECISION_LOG.md
    SESSION_NOTES/
      YYYY-MM-DD-session.md
```

If the project is pushed to GitHub, its memory travels with that repo.

## 7. Core Workflows

### Workflow A - Save / Compress

Use when the user wants to preserve the current state before tokens run out, before stopping for the night, after meaningful progress, after a commit, or before switching tools.

Trigger phrases include:

```text
use ContextKit
save project memory
compress context
create handoff note
prepare next agent handoff
update project context
```

Expected result:

- repo memory files are created or updated
- the decision log changes only if a real decision was made
- today's session note captures changed files, known issues, and next steps
- the agent reports a short handoff summary and suggested next prompt

### Workflow B - Resume / Retrieve

Use when a new session or different agent needs to understand the latest project state before editing.

Trigger phrases include:

```text
resume with ContextKit
load project memory
read project context
catch me up
what is the latest project state?
start from the last handoff
```

Expected result:

- the agent reads repo-local memory files
- the agent inspects current Git state
- the agent reports the current phase, branch, locked decisions, risks, relevant files, and safest next task
- no files are edited unless the user asks to save/compress context

## 8. Files To Inspect

For both workflows, inspect Git state first:

```text
git rev-parse --show-toplevel
git status --short --branch
git log --oneline -10
```

For Save / Compress, inspect changed files or targeted diffs as needed before writing memory updates.

Read these files if they exist:

```text
AGENTS.md
docs/PROJECT_CONTEXT.md
docs/DECISION_LOG.md
latest file in docs/SESSION_NOTES/
```

`docs/RELEASE_CHECKLIST.md` or project-specific planning docs may be read when present, but they are not required ContextKit files.

## 9. Files To Create Or Update

ContextKit may create or update:

```text
AGENTS.md
docs/PROJECT_CONTEXT.md
docs/DECISION_LOG.md
docs/SESSION_NOTES/YYYY-MM-DD-session.md
```

Do not update production code during ContextKit save/resume unless the user explicitly asks for production-code changes.

## 10. Memory File Templates

### PROJECT_CONTEXT.md

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

Rules:

- keep it under roughly 80 lines
- keep only current truth
- prefer bullets over paragraphs
- remove outdated items
- do not include long code details
- do not duplicate the session note

### DECISION_LOG.md

```md
# PROJECT_NAME Decision Log

## YYYY-MM-DD - Decision title

Decision: ...
Reason: ...
Status: Locked / Accepted / Deferred / Rejected
Reopen only if: ...
```

Rules:

- update only when a real decision was made
- do not log tiny implementation details
- log decisions that prevent future rework
- do not mark a decision as locked unless user confirmation or repo evidence supports it

### SESSION_NOTES/YYYY-MM-DD-session.md

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

Rules:

- this is the practical handoff note
- include the exact next prompt the user can copy/paste
- keep it compressed and factual
- use `needs verification` when evidence is incomplete

## 11. AGENTS.md Addition

If `AGENTS.md` exists, append this section if missing. If it does not exist, create it with this section and any safe project-specific instructions that can be inferred from repo files.

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

## 12. Evidence And Safety Rules

- Use Git state and repo files as the source of truth.
- Do not invent progress.
- Do not mark something as done unless Git, file content, or the user confirms it.
- If unsure, write `needs verification`.
- If a memory file appears outdated, update or remove stale points.
- Do not expose secrets, tokens, credentials, private keys, or sensitive customer data.
- Do not run destructive Git commands.
- Do not delete branches.
- Do not rewrite commit history.

## 13. Final Response Requirements

After Save / Compress, reply with:

1. files updated
2. current project phase
3. main locked decisions
4. known risks
5. next recommended task
6. suggested next prompt

After Resume / Retrieve, reply with:

1. current project phase
2. active branch and Git state
3. latest handoff summary
4. locked decisions
5. known risks
6. safest next task

Keep both responses short.

## 14. Acceptance Criteria

The implementation is complete when:

- `metis-lab/skills/contextkit/SKILL.md` exists and has valid skill frontmatter.
- `metis-lab/skills/contextkit/agents/openai.yaml` exists and follows the local skill metadata format.
- this PRD exists as the generic ContextKit reference.
- the skill clearly separates Save / Compress and Resume / Retrieve.
- the skill instructs agents to inspect Git state before writing memory files.
- the skill forbids production-code edits during save/resume unless explicitly requested.
- the skill defines repo-local memory files for each project.
- `metis-lab/scripts/sync_skills.sh` includes `contextkit` in reusable skill sync.

## 15. Suggested Commit Message

```text
docs: add ContextKit project memory skill
```
