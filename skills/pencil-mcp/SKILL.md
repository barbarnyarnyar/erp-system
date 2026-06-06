---
name: "pencil-mcp"
description: "Use when working with Pencil.dev and Pencil MCP tools to inspect .pen files, edit designs, compose screens with design-system components, validate layout/screenshot quality, or generate implementation-ready UI updates. Trigger for requests about Pencil MCP workflows, .pen editing, design automation, and design-to-dev execution in Codex."
---

# Pencil MCP Workflow

Current local skill version: `v1.5.0`  
Last updated: `2026-03-18`

Follow this workflow when editing or reviewing .pen files through Pencil MCP.

## 1) Start with editor context

1. Call `mcp__pencil__get_editor_state` first.
2. Identify: active file, current selection, and whether a design system exists.
3. If no document is open and work must start from scratch, call `mcp__pencil__open_document`.

### Codex in Antigravity: MCP recovery signals

When Pencil MCP is used from Codex inside Antigravity, distinguish transport failures from editor-state failures before doing anything drastic.

- `WebSocket not connected to app: antigravity`
  - Treat as a bridge/app connection problem.
  - Ask the user to restart or refocus Antigravity and retry.
- `A file needs to be open in the editor`
  - Treat as an active-editor/session-routing problem, not a missing `.pen` file by default.
  - If another Antigravity agent can see the file, assume Codex is bound to the wrong or stale editor context.

Preferred recovery order:
1. Keep Antigravity as the active Pencil environment.
2. Reduce competing editor contexts when possible, especially regular VS Code if Codex/Pencil are installed in both places.
3. Ensure the target `.pen` tab is directly focused in the active Antigravity session.
4. Retry `mcp__pencil__get_editor_state`.
5. Only after that consider deeper troubleshooting or reinstall assumptions.

## 2) Discover structure before editing

1. Use `mcp__pencil__batch_get` to read top-level frames and reusable components.
2. Read only the necessary depth (`readDepth` low by default).
3. Prefer one combined `batch_get` request for multiple IDs/patterns.

### Pencil as intermediate representation

When Pencil is being used as the design source for implementation work, treat the selected Pencil section as:

- the visual source of truth
- the intended DOM object structure for that section

Do not treat Pencil as image-only reference.

Before editing code for a section, extract and state:

1. wrapper hierarchy
2. sibling order
3. content grouping
4. semantic intent of each group
5. responsive structural differences, if any

The code pass should then preserve both:

- visual parity
- structure parity

If a section looks close visually but the DOM grouping is materially different from Pencil, treat that as incomplete parity.

## 3) Pull targeted guidance

1. Choose the closest guideline topic (`design-system`, `web-app`, `mobile-app`, `landing-page`, `table`, `code`) via `mcp__pencil__get_guidelines`.
2. Use style guide helpers only when the task needs visual direction from scratch.

## 4) Apply edits safely

1. Use `mcp__pencil__batch_design` with small batches (up to ~25 operations).
2. Always bind `I`, `C`, and `R` operations to variable names.
3. Prefer component instances (`ref`) over rebuilding reusable UI manually.
4. Use `R` for child replacement inside instances; do not set `children` with `U`.
5. When copying nodes and customizing descendants, put descendant overrides in the `C(..., {"descendants": ...})` call.

### Live content update rule

When the user asks to change wording, labels, headings, or other on-screen content in an active `.pen` file:

1. Find the live node IDs with `batch_get` or current selection context.
2. Update the active editor file through Pencil MCP operations such as `U("nodeId",{content:"..."})`.
3. Treat this MCP update as the primary source of truth for the visible screen.
4. Only mirror the same change into a repo-tracked `.pen` copy if that project keeps a second versioned copy.

Do **not** use raw file patching as the primary method for visible screen text/content changes when Pencil MCP can address the live node directly.

Raw `.pen` file patching is acceptable only for:

- repo-copy mirroring after the live MCP update
- bulk maintenance when no active editor context is needed
- recovery/fallback when MCP editing is unavailable

Avoid the two-step anti-pattern:

1. patch `.pen` JSON behind the editor first
2. then separately update the live screen through MCP

That creates unnecessary drift risk and duplicate work.

## 5) Validate each pass

1. Run `mcp__pencil__snapshot_layout` (use `problemsOnly=true` when checking clipping/overlap quickly).
2. Capture visual output with `mcp__pencil__get_screenshot` for key frames/components.
3. Iterate until layout and visual checks are clean.

### Responsive logo sync rule

When a selected desktop header/footer logo treatment must propagate across responsive canvases:

1. Use the desktop instance as the treatment source of truth.
2. Propagate artwork/treatment only.
3. Preserve each breakpoint's local width, height, and placement.
4. Validate each responsive header/footer with Pencil screenshots before any code sync begins.

Do not enlarge the desktop source to match other breakpoints. The desktop treatment is the baseline.

### Design-to-code parity loop

For section-by-section parity work:

1. Read the selected Pencil subtree for the active section.
2. Capture a Pencil screenshot for that section.
3. Extract structure, text, colors, spacing, corner radii, and responsive differences.
4. Write or state the intended DOM tree for that section before editing code.
5. Compare those directly against the matching HTML/CSS/JS implementation.
6. Fix the active section only.
7. Capture browser screenshots after the code pass.
8. Do not move to the next section until the current section is visually clean and structurally aligned.

### Explicit token parity checks

For synced sections, check these explicitly instead of assuming they carried over:

- wrapper hierarchy and grouping parity
- DOM order parity
- spacing-token usage
- corner-radius-token usage
- wrapper-stack structure
- responsive breakpoint differences

### Native behavior rule

When a browser interaction does not require JavaScript state, prefer native HTML behavior over JS wrappers.

Examples:

- file downloads should prefer `<a href ... target="_blank" rel="noopener noreferrer">`
- external links should prefer native anchors

Do not keep a JS-triggered action when native markup is clearer, faster, and more reliable.

### Launch polish anti-patterns to avoid

From recent Trivyst work, these are recurring failure modes:

1. updating HTML without the matching CSS pass
2. updating desktop Pencil only and leaving responsive variants stale
3. patching repo `.pen` copies before updating the live active Pencil canvas
4. trusting browser refresh instead of screenshot parity
5. treating spacing/radius as visual afterthoughts instead of token checks
6. leaving stale helper CSS in place after section structure changes
7. using JS for interactions that should be native links/buttons
8. assuming the active Pencil tab is saved without an explicit save handshake

## 5.5) Save handshake (required)

1. After any create/update in a .pen file, explicitly ask the user to save the file in the editor.
2. Also offer to save on the user's behalf by copying the active .pen file to a target path the user confirms.
3. If the user wants assistant-side save, perform it immediately and report the exact output path.
4. Do not assume edits are safely persisted until this handshake is complete.

### Specialized Visual Directions: Neomorphism (Embossed)

When a task requires an "embossed", "soft UI", or "neomorphic" look:

1.  **Uniform Surface**: Component `fill` MUST exactly match the parent frame `fill`.
2.  **Shadow Physics**:
    *   **Extruded (Outset)**: Use two outer shadows. One light (top-left) and one dark (bottom-right).
    *   **Inset (Inward)**: Use two inner shadows. One light (bottom-right) and one dark (top-left).
3.  **Softness**: Use high corner radii (e.g., 20px-48px) and soft blur values (e.g., 10px-20px) to maintain the "molded" physical appearance.
4.  **Interaction**: Flip from Extruded to Inset to simulate a physical button press.

## 6) Design-to-dev handoff

1. Pull variables with `mcp__pencil__get_variables` when generating code.
2. Keep token names and spacing/typography scales intact in implementation output.
3. Provide explicit file references and concrete changes for engineering follow-through.

## 6.5) Documented Design ↔ Code workflow

Per Pencil docs, the intended design-to-code workflow is explicit, not magical:

1. Keep the `.pen` file inside the project workspace alongside the codebase.
2. Save the `.pen` file explicitly before handoff/sync steps.
3. Use Pencil as the design source of truth for the current pass.
4. Then run a deliberate sync step:
   - Design → Code for new work
   - Code → Design for existing work that must first be imported
5. Refine the generated/implemented code after the initial sync.

Do not present normal HTML/CSS/JS implementation work as an automatic Pencil converter unless an explicit Pencil code-generation/export step is actually being used.

### Workflow recommendations to follow

For new features/pages:

1. Design in Pencil first.
2. Generate or implement the initial code from that design.
3. Refine the code manually.
4. Re-check against Pencil screenshots.

For existing features/pages:

1. Import or reconstruct the current code/component into Pencil context first.
2. Make the design change in Pencil.
3. Sync changes back to code.

### Variables and token sync

Pencil variables should be treated as the design-side equivalent of CSS custom properties.

1. Keep meaningful tokens mirrored between Pencil and code:
   - color
   - typography
   - spacing
2. Prefer semantic token names over one-off raw values.
3. When working with an existing design system, add project-specific token layers in parallel instead of mutating the base system by default.

## Guardrails

- Keep context lean; avoid deep full-tree reads unless necessary.
- Preserve existing design system language when editing an established product.
- When creating new marketing/landing surfaces, pick a deliberate visual direction and avoid default generic styling.
- For large redesigns, work section-by-section and verify after each section.

For detailed execution patterns and Codex/Antigravity recovery notes, see `references/workflows.md`.

For local version history, see `references/changelog.md`.
