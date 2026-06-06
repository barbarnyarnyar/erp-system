# Pencil MCP Workflows

## Rapid Fix

1. `get_editor_state`
2. `batch_get` for affected frame + reusable components
3. `batch_design` minimal patch
4. `snapshot_layout` + `get_screenshot`

## New Section in Existing App Screen

1. `batch_get` existing screen container and relevant reusable components
2. `get_guidelines` with `design-system` or `web-app`
3. `batch_design` section scaffold first, content second
4. Validate with screenshot and layout checks

## Net-new Screen

1. `find_empty_space_on_canvas`
2. Insert top-level frame
3. Build hierarchy in multiple `batch_design` calls
4. Validate after each structural pass

## Codex in Antigravity Recovery

Use this when Pencil MCP behaves inconsistently inside Codex on Antigravity.

### Failure signatures

- `WebSocket not connected to app: antigravity`
  - Bridge/app session problem.
- `A file needs to be open in the editor`
  - Active editor is not exposed to the current Codex MCP session.
- Another Antigravity agent can access the same `.pen` file, but Codex cannot
  - Strong signal that Pencil is fine and Codex is attached to the wrong or stale editor context.

### Recovery sequence

1. Keep one Antigravity session as the primary Pencil environment.
2. Focus the target `.pen` tab directly in Antigravity.
3. Reduce competing editor contexts when practical, especially regular VS Code if Pencil/Codex are installed there too.
4. Retry `get_editor_state`.
5. If still stale, restart Antigravity and retry before assuming reinstall is needed.

### Interpretation rule

Do not jump straight to reinstall.

First decide whether the failure is:
- connection-layer (`WebSocket not connected`)
- editor-state-layer (`A file needs to be open in the editor`)

In the observed Codex/Antigravity case, the successful fix was environment cleanup plus Antigravity restart, not reinstalling Pencil or replacing the `.pen` file.
