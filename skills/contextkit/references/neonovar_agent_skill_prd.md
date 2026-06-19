# PRD — NEONOVAR Agent Skill: Project Context Compression

## 1. Product Summary

Build a lightweight repo-native agent skill that allows Codex or another AI coding agent to periodically compress NEONOVAR project progress into structured Markdown memory files.

The purpose is to preserve project continuity across token limits, interrupted night sessions, AI tool switching, and multi-day gaps.

This skill should not build product features. It should inspect project state, recent commits, current changes, and existing memory docs, then update compressed context files so the next agent session can resume quickly.

---

## 2. Problem Statement

The NEONOVAR portfolio project is being developed through AI-assisted coding, visual design iteration, and Git-based experimentation. Work often happens in late-night sessions. Sometimes the AI session ends because of tool limits, token limits, fatigue, or time gaps.

When this happens, the next session loses important context:

- What was already decided
- What should not be reopened
- What branch is active
- What changed recently
- What remains broken
- What the next safe task should be
- Which files matter most

Without repo-native memory, each new session wastes time reconstructing state from chat history, Git graph, or human memory.

---

## 3. Goal

Create a simple agent skill that updates project memory every 60–90 minutes or whenever the user asks for context compression.

The skill should make the repo itself the source of continuity.

---

## 4. Non-Goals

This skill must not:

- Implement new website features
- Redesign sections
- Refactor production code
- Add new dependencies
- Rewrite large documentation files unnecessarily
- Invent progress that cannot be verified
- Mark incomplete work as done
- Replace Git commits, issues, or release tags

---

## 5. Target User

Primary user: Aung Shwe Oo, solo designer-builder working on the NEONOVAR portfolio site with AI-assisted coding.

Secondary user: Any AI coding agent that needs to resume the project after context loss.

---

## 6. Core Use Cases

### Use Case 1 — Hourly Context Compression

The user has been working for about 60–90 minutes and wants to preserve current progress.

User prompt example:

```txt
Use the Update Project Context Memory skill. Compress current project state. Do not code.
```

Expected result:

- `docs/PROJECT_CONTEXT.md` is updated
- `docs/DECISION_LOG.md` is updated only if a real decision was made
- Today’s session note is created or updated
- The agent reports what changed and what should happen next

### Use Case 2 — End-of-Night Handoff

The user is stopping work for the night and wants tomorrow’s AI session to restart cleanly.

Expected result:

- Today’s session note includes changed files, decisions, known issues, and the next recommended prompt
- Current context file reflects the latest project phase
- Old or outdated context is removed

### Use Case 3 — Agent Tool Switching

The user is switching from Codex to another agent, or from one chat session to another.

Expected result:

- New agent can read repo memory files and continue without needing full chat history

### Use Case 4 — Production Candidate Freeze

The project is in V1 production candidate mode. The agent should help prevent unnecessary reopening of design decisions.

Expected result:

- Locked decisions are preserved in `DECISION_LOG.md`
- `PROJECT_CONTEXT.md` clearly separates allowed fixes from forbidden rework

---

## 7. Recommended File Structure

The implementation should create or maintain this structure:

```txt
AGENTS.md
docs/
  PROJECT_CONTEXT.md
  DECISION_LOG.md
  RELEASE_CHECKLIST.md
  SESSION_NOTES/
    YYYY-MM-DD-session.md
.agent-skills/
  update-project-context.md
```

If `.agent-skills/` already exists, use it. If another agent-skill directory already exists in the project, follow the existing project convention.

---

## 8. Files to Read Before Updating

When the skill is invoked, the agent should read or inspect:

```txt
AGENTS.md
docs/PROJECT_CONTEXT.md
docs/DECISION_LOG.md
docs/RELEASE_CHECKLIST.md
latest file in docs/SESSION_NOTES/
git status
git log --oneline -10
recently changed files from git diff / git status
```

If any file does not exist, the agent may create it using the templates in this PRD.

---

## 9. Files to Update

The skill should update:

```txt
docs/PROJECT_CONTEXT.md
docs/DECISION_LOG.md
docs/SESSION_NOTES/YYYY-MM-DD-session.md
```

It may also update `AGENTS.md` once to add the project memory protocol if it is missing.

It should not update production code unless explicitly asked.

---

## 10. Skill Invocation Names

The skill should respond to prompts such as:

```txt
update project context
compress context
save progress memory
create handoff note
use project context skill
use memory compression skill
prepare next agent handoff
```

---

## 11. Agent Skill File Content

Create this file:

```txt
.agent-skills/update-project-context.md
```

Recommended content:

```md
# Skill: Update Project Context Memory

## Purpose

Use this skill to compress the current project state into repo-native memory files so future AI sessions can resume quickly after token limits, context loss, tool switching, or time gaps.

This skill does not implement new features unless explicitly asked. Its job is to inspect recent progress, summarize decisions, and update project continuity notes.

## When to use

Use this skill:

- Every 60–90 minutes during active work
- Before stopping for the night
- After a meaningful commit
- After finishing a feature, bugfix, design direction, or refactor
- Before switching from one AI agent/tool to another
- When the user says: “update project context”, “compress context”, “save progress memory”, “handoff note”, or “use project context skill”

## Files to read first

Read these files if they exist:

- `AGENTS.md`
- `docs/PROJECT_CONTEXT.md`
- `docs/DECISION_LOG.md`
- `docs/RELEASE_CHECKLIST.md`
- latest file in `docs/SESSION_NOTES/`
- current `git status`
- recent commits from `git log --oneline -10`

Also inspect recently changed files from Git before writing updates.

## Files to update

Update these files as needed:

- `docs/PROJECT_CONTEXT.md`
- `docs/DECISION_LOG.md`
- `docs/SESSION_NOTES/YYYY-MM-DD-session.md`

Do not create excessive documentation. Keep updates compressed, factual, and useful for the next AI session.

## Output goals

After running this skill, the repo should clearly answer:

1. What phase is the project in?
2. What branch is active?
3. What changed recently?
4. What is already locked and should not be reopened?
5. What risks or bugs remain?
6. What should the next session do first?
7. What files are most relevant for the next session?

## Update rules

### PROJECT_CONTEXT.md

Keep this file short. It should be the current high-level project state, not a diary.

Use this structure:

```md
# NEONOVAR Project Context

## Current phase
...

## Current branch
...

## Current status
...

## Site sections
- ...
- ...
- ...

## Current priorities
1. ...
2. ...
3. ...

## Do not reopen
- ...
- ...
- ...

## Known risks
- ...
- ...
- ...

## Next best task
...
```

Rules:

- Remove outdated items
- Keep only the current truth
- Prefer bullets over paragraphs
- Keep it under roughly 80 lines
- Do not include long code details
- Do not duplicate the session note

### DECISION_LOG.md

Update only when a real decision was made.

Use this format:

```md
## YYYY-MM-DD — Decision title

Decision: ...
Reason: ...
Status: Locked for V1 / Accepted / Deferred / Rejected
Reopen only if: ...
```

Rules:

- Do not log tiny implementation details
- Log decisions that prevent future rework
- Be explicit about what should not be reopened
- Do not mark something as locked unless the user or recent work clearly confirms it

### SESSION_NOTES/YYYY-MM-DD-session.md

Create or update today’s session note.

Use this format:

```md
# Session Note — YYYY-MM-DD

## Branch
...

## Session goal
...

## Changed
- ...
- ...

## Files touched
- ...
- ...

## Decisions made
- ...
- ...

## Known issues
- ...
- ...

## Next session should start with
1. ...
2. ...
3. ...

## Suggested next prompt
...
```

Rules:

- This is the handoff note
- It should help a new AI agent resume without full chat history
- Include the exact next prompt the user can copy/paste
- Keep it practical and compressed

## Evidence rules

- Use Git state and files as the source of truth
- Do not invent progress
- Do not mark something as done unless Git, file content, or the user confirms it
- If unsure, write “needs verification”
- If a file appears outdated, update or remove stale points

## Safety rules

- Do not change production code while using this skill unless the user explicitly asks
- Do not add dependencies
- Do not run destructive Git commands
- Do not delete branches
- Do not rewrite commit history
- Do not expose secrets or credentials in docs

## Final response to user

After updating, reply with:

1. Files updated
2. Current project phase
3. Main locked decisions
4. Known risks
5. Next recommended task
6. Suggested next prompt

Keep the response short.
```

---

## 12. AGENTS.md Addition

If `AGENTS.md` exists, append this section if missing:

```md
## Project memory protocol

This repo uses compressed project memory to preserve continuity across AI sessions.

Before starting meaningful work, read:

- `docs/PROJECT_CONTEXT.md`
- `docs/DECISION_LOG.md`
- latest file in `docs/SESSION_NOTES/`

When asked to “update project context”, “compress context”, “save progress memory”, or “handoff note”, use the project context memory process:

1. Inspect `git status`.
2. Inspect `git log --oneline -10`.
3. Review changed files if needed.
4. Update `docs/PROJECT_CONTEXT.md`.
5. Update `docs/DECISION_LOG.md` only for real decisions.
6. Create or update today’s session note in `docs/SESSION_NOTES/`.
7. Do not change production code during context compression unless explicitly asked.
```

If `AGENTS.md` does not exist, create it and include this section plus any existing project-specific instructions the agent can safely infer from the repo.

---

## 13. Initial PROJECT_CONTEXT.md Template

Create this file if missing:

```txt
docs/PROJECT_CONTEXT.md
```

Initial content:

```md
# NEONOVAR Project Context

## Current phase
V1 production candidate freeze.

## Current branch
site/v1-production-candidate

## Current status
NEONOVAR V1 is being stabilized. Exploration should be minimized. Focus on final fixes, responsive checks, form behavior, image optimization, and release readiness.

## Site sections
- Landing page mostly stable.
- Contact section implemented with modal/Formspree flow; needs verification if not recently tested.
- About page layout and copy fixes mostly done; needs final responsive check.
- Teaching page hero may still need refinement.
- Parallax groundwork exists but should not be expanded unless required for a defect.

## Current priorities
1. Fix bugs.
2. Tighten mobile/responsive issues.
3. Finalize Teaching page.
4. Optimize images.
5. Prepare `v1.0-rc1` tag.

## Do not reopen
- Overall dark visual language.
- Contact section core direction.
- Index grid/modal concept.
- General typography direction.

## Known risks
- Animation performance on mobile.
- Over-polishing visual variants.
- Form validation and success/error state.
- Image weight.

## Next best task
Run a V1 production-candidate verification pass: build, responsive check, contact form test, console error check, and image-weight review.
```

---

## 14. Initial DECISION_LOG.md Template

Create this file if missing:

```txt
docs/DECISION_LOG.md
```

Initial content:

```md
# NEONOVAR Decision Log

## 2026-05-14 — V1 production candidate freeze

Decision: Treat the current site direction as V1 production candidate. Focus on defects, responsive fixes, form behavior, image optimization, and release readiness.
Reason: The visual system and main site direction are stable enough; continued exploration risks delaying release.
Status: Locked for V1
Reopen only if: A core section is broken, unusable, inaccessible, or clearly off-brand.

## 2026-05-14 — Dark AI-era visual direction retained

Decision: Keep the dark navy/black, cyan/violet glow, premium AI-era editorial direction.
Reason: This is the established NEONOVAR identity and should not be reworked during V1 freeze.
Status: Locked for V1
Reopen only if: Performance or readability fails.
```

---

## 15. Initial Session Note Template

Create today’s note if missing:

```txt
docs/SESSION_NOTES/YYYY-MM-DD-session.md
```

Use the current date from the machine or Git environment.

Template:

```md
# Session Note — YYYY-MM-DD

## Branch
site/v1-production-candidate

## Session goal
Compress project state for AI-assisted continuity.

## Changed
- Added or updated project memory files.
- Added agent skill for context compression.

## Files touched
- `.agent-skills/update-project-context.md`
- `docs/PROJECT_CONTEXT.md`
- `docs/DECISION_LOG.md`
- `docs/SESSION_NOTES/YYYY-MM-DD-session.md`
- `AGENTS.md` if project memory protocol was missing

## Decisions made
- Use repo-native Markdown memory as the continuity layer for AI-assisted work.
- Use session notes instead of relying on chat history.

## Known issues
- Needs verification against current Git state.
- Needs future updates every 60–90 minutes or before stopping work.

## Next session should start with
1. Read `AGENTS.md`.
2. Read `docs/PROJECT_CONTEXT.md`.
3. Read `docs/DECISION_LOG.md`.
4. Read the latest file in `docs/SESSION_NOTES/`.
5. Inspect `git status` and `git log --oneline -10`.

## Suggested next prompt
Read AGENTS.md, docs/PROJECT_CONTEXT.md, docs/DECISION_LOG.md, and the latest docs/SESSION_NOTES file. Then inspect git status and git log --oneline -10. Tell me the current phase, what is locked, what remains, and the safest next task. Do not edit files yet.
```

---

## 16. Acceptance Criteria

The implementation is complete when:

- `.agent-skills/update-project-context.md` exists and contains the skill instructions.
- `docs/PROJECT_CONTEXT.md` exists and reflects the current high-level project state.
- `docs/DECISION_LOG.md` exists and includes at least the V1 freeze decision.
- `docs/SESSION_NOTES/` exists.
- A dated session note exists for the implementation day.
- `AGENTS.md` contains the project memory protocol, either newly added or confirmed existing.
- The skill clearly instructs agents not to code while compressing context.
- The skill tells agents to inspect Git state before updating docs.
- The skill output includes a short handoff summary and suggested next prompt.

---

## 17. Quality Bar

The implementation should be:

- Small
- Markdown-only
- Easy to understand
- Easy for Codex to follow
- Safe for production-candidate work
- Free of unnecessary automation or dependencies
- Useful even if another AI agent reads it later

---

## 18. Suggested Implementation Steps for Codex

1. Inspect repo root.
2. Check if `AGENTS.md` exists.
3. Check if `docs/` exists.
4. Check if `.agent-skills/` exists.
5. Create missing directories.
6. Add `.agent-skills/update-project-context.md`.
7. Create or update `docs/PROJECT_CONTEXT.md`.
8. Create or update `docs/DECISION_LOG.md`.
9. Create `docs/SESSION_NOTES/YYYY-MM-DD-session.md`.
10. Add project memory protocol to `AGENTS.md` if missing.
11. Run `git status`.
12. Report changed files and next recommended prompt.

---

## 19. Suggested Commit Message

```txt
docs: add project context compression skill
```

---

## 20. References Behind the Approach

This PRD follows the general pattern of using repo-native AI instruction and memory files. OpenAI Codex supports project guidance through `AGENTS.md`, including repository-level instructions and working agreements. Claude Code similarly supports project memory through context files such as `CLAUDE.md`, where concise instructions are loaded as context at the start of conversations.

The NEONOVAR version intentionally stays tool-agnostic by using Markdown files inside the repo rather than relying only on one AI tool’s memory system.

