# Pencil MCP Local Changelog

This changelog tracks local updates to the shared `pencil-mcp` skill in the Antigravity/Codex environment.

## v1.5.1 — 2026-03-21

### Type

- Documentation-only follow-up to `v1.5.0`
- No change to the main skill body yet

### Findings recorded

From the Trivyst launch follow-through after `v1.5.0`, these workflow gaps became clearer:

1. phone-width checks were not enough
   - iPad Mini portrait
   - iPad Air portrait
   - smaller tablet portrait widths
   must be treated as required responsive checkpoints
2. section parity was still being closed too loosely
   - visuals were improved in batches
   - but structure, tokens, spacing, and responsive validation were not always closed before moving on
3. some repeated fixes were actually system problems first
   - missing radius token normalization
   - missing font-size token normalization
   - missing line-height token normalization
   - semantic list markup drifting into `p + br + bullet` patterns
4. vertical rhythm needed to be treated as a first-class parity check
   - section padding
   - internal stack gaps
   - card padding
   - CTA spacing
   - hero vertical centering
5. header/nav must be audited as a cross-breakpoint system
   - shell radius
   - shell height
   - shell padding
   - logo scale
   - toggle size and bar treatment
   - open/close icon treatment
6. code content can become newer than Pencil during launch polish
   - when that happens, reverse-sync back into Pencil early instead of letting wording drift grow
7. MCP/toolchain health must be checked before parity work starts
   - Pencil MCP
   - Browser MCP
   - screenshot/checkpoint path

### Why this note was added

The generic skill already became stricter in `v1.5.0`, but the last few days showed that real launch work fails not only on DOM structure drift. It also fails on:

- incomplete breakpoint coverage
- token-system drift
- weak section completion gates
- delayed semantic cleanup
- late discovery of toolchain issues

This `v1.5.1` entry records those operational findings before deciding which parts should be promoted into the shared skill body later.

### Candidate changes for a future skill update

Potential future promotion into the main skill:

1. required responsive checkpoint matrix
2. section completion gate
3. system normalization pre-pass
4. explicit vertical scale audit
5. header/nav cross-breakpoint audit
6. reverse-sync rule when HTML becomes the newer content source
7. lightweight anticipatory `wizard mode` questioning at major checkpoints

## v1.5.0 — 2026-03-18

### Added

- Explicit rule that Pencil should be treated as the intermediate representation of both visual design and intended DOM object structure during implementation passes.
- Required step to state the intended DOM tree for each selected Pencil section before editing code.
- Stronger parity checklist covering wrapper hierarchy and DOM order, not just visuals.
- Native behavior rule preferring semantic HTML for downloads/links when JavaScript is unnecessary.
- Anti-pattern checklist based on Trivyst launch polishing failures.

### Why this update was made

During the Trivyst launch pass, several problems repeated even after visuals were close:

1. section structure drift between Pencil and HTML
2. stale responsive variants in Pencil before code sync
3. old helper CSS continuing to influence layout after structure changes
4. JS-wrapped download behavior where a native link would have been cleaner
5. repeated confusion about whether refresh problems were actually parity problems

The workflow needed to become stricter about using Pencil as the intermediate structure source, not just a screenshot reference.

### Practical rule going forward

1. Read the selected Pencil section.
2. State the intended DOM hierarchy from that section.
3. Sync live Pencil first, then repo mirror if needed, then code.
4. Validate visual and structural parity section-by-section.
5. Prefer native browser behavior when no JS state is required.

## v1.4.0 — 2026-03-18

### Added

- Responsive logo sync rule for propagating selected desktop logo treatment across responsive Pencil canvases without enlarging the desktop source.
- Official section-by-section design-to-code parity loop for Pencil-guided implementation work.
- Explicit token parity checks for spacing and corner radius.
- Validation rule that browser screenshots must be checked after each section pass before moving on.

### Why this update was made

During Trivyst final polishing, responsive header/footer wordmarks and wrapper-level spacing drift kept reappearing because design updates were not being validated through one consistent loop. The workflow needed a stricter section-by-section method that starts from Pencil truth, checks browser output, and treats spacing/radius as first-class parity checks.

### Practical rule going forward

1. Normalize shared responsive logo treatments in Pencil first.
2. Work one section at a time.
3. Compare Pencil screenshots against matching HTML/CSS/JS.
4. Verify spacing-token and corner-radius-token parity explicitly.
5. Capture browser screenshots before advancing to the next section.

## v1.3.0 — 2026-03-18

### Added

- Official Pencil docs guidance for the Design ↔ Code workflow.
- Explicit rule that normal Pencil-guided implementation should not be described as an automatic converter unless an actual code-generation/export step is run.
- Repo-first file organization guidance: keep `.pen` files in the same project workspace as the codebase.
- Documented recommendation to use Design → Code for new work and Code → Design first for existing work.
- Stronger token-sync rule between Pencil variables and CSS custom properties.

### Why this update was made

During Trivyst implementation, Pencil MCP was being used correctly for design context and live design editing, but the HTML pass was still a manual translation. The workflow needed to be documented more precisely so future work is described and executed as an explicit sync process rather than an implied one-click converter.

### Documentation basis

Aligned to Pencil docs covering:

1. AI Integration
2. Design ↔ Code
3. Import & Export

### Practical rule going forward

1. Keep `.pen` in the repo workspace.
2. Use Pencil MCP to establish or update the live design state.
3. Choose the right sync direction:
   - Design → Code for new work
   - Code → Design → Code for existing work
4. Mirror tokens/variables between Pencil and CSS deliberately.
5. Describe manual implementation honestly as Pencil-guided sync, not automatic conversion.

## v1.2.0 — 2026-03-17

### Added

- Live content update rule for wording, labels, and text changes inside active `.pen` files.
- Explicit requirement to update on-screen node IDs through Pencil MCP first when the user expects the live canvas to change.
- Guidance that raw `.pen` JSON patching should be secondary, used only for repo-copy mirroring, maintenance, or fallback.

### Why this update was made

During Trivyst homepage refinement, a wording change was first patched directly in the `.pen` file and then updated again through Pencil MCP so the live screen would actually reflect it. That two-step pattern was unnecessary and created avoidable drift risk between the editor-visible design and the versioned file copy.

### Correct update order confirmed

1. Identify the live text nodes in the active editor file.
2. Update those nodes through Pencil MCP operations.
3. Mirror the same change into any repo-tracked `.pen` copy only if the project keeps one.
4. Avoid behind-the-editor file patching as the primary path for visible screen updates.

## v1.1.0 — 2026-03-16

### Added

- Codex-in-Antigravity recovery guidance for Pencil MCP failures.
- Explicit interpretation of:
  - `WebSocket not connected to app: antigravity`
  - `A file needs to be open in the editor`
- Recovery sequence for likely stale or wrong editor binding.
- Rule to prefer Antigravity as the active Pencil environment when Codex and Pencil are installed in both Antigravity and regular VS Code.

### Why this update was made

During Trivyst design work, Pencil MCP initially failed from Codex even though the `.pen` file existed and another Antigravity agent could access it. The issue was traced to session/editor-state routing rather than a broken file or required reinstall.

### Working recovery pattern confirmed

1. Treat `WebSocket not connected` as a bridge/app connection issue.
2. Treat `A file needs to be open in the editor` as an active-editor/session issue by default.
3. Reduce competing editor contexts when possible.
4. Focus the target `.pen` tab directly in Antigravity.
5. Retry `get_editor_state`.
6. Restart Antigravity before assuming reinstall is necessary.

## v1.0.0 — Previous baseline

Initial local workflow version:

- basic Pencil MCP workflow
- design editing sequence
- save handshake rule
- design-to-dev handoff guidance
