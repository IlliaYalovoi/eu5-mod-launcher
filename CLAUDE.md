# AI AGENT RULES (MANDATORY)

## CORE BEHAVIOR
- NO preamble, NO explanations, NO summaries
- DO NOT describe actions — perform them
- NO follow-up questions unless BLOCKED by ambiguity
- OUTPUT ONLY:
   - file changes
   - unresolved issues (if any)
- TERMINATE immediately after output

VIOLATION = INVALID RESPONSE

---

## ENVIRONMENT
- OS: Linux (WSL2) | Shell: Bash
- Stack: Go (Wails), Vue 3 + TypeScript

---

## PROJECT
- Desktop app for managing mods across multiple Paradox games titles.

## EXECUTION RULES
- Edit in-place on main branch (NO worktrees)
- Plan internally, execute in batches
- DO NOT interleave planning and execution
- On failure:
   - state failure
   - state cause
   - state requirement
   - max 2 retries

---

## FILE ACCESS
- ≤100 lines → read full
- >100 lines → read 1–50, then targeted reads
- Full read REQUIRED before full rewrite
- NEVER re-read unchanged files
- Go → read relevant function only (unless rewrite)
- Vue → read sections separately if large

---

## TOOLING RULES
- Prefer native tools (Read/Glob/Edit) over Bash
- Bash only when necessary

### REQUIRED
- Search: `rg`
- File discovery: `fd`

### BANNED (ABSOLUTE)
- `ls`
- `grep`
- `find`
- `sed -i` (for code edits)
- re-reading files in context

USE:
- Glob → directory listing
- Read → existence check
- Edit → code changes

ANY USE OF BANNED COMMANDS = FAILURE

---

## TOOL EXECUTION DISCIPLINE
- Batch operations:
   - `mkdir -p a b c`
   - multiple writes in parallel
- NEVER verify with `ls`
- NEVER grep→sed pattern
- ALWAYS:
   - plan operations first
   - execute in batch
- ONE verification pass at end (if needed)
- NO step-by-step verification

---

## PARALLELISM
- Independent reads → parallel
- Independent writes → parallel
- NO sequential calls without dependency

---

## GO RULES
- Match local code style
- Self-documenting names (minimal comments)
- No partial states across files
- Validate:
   - call sites
   - interfaces
   - types
- Imports: std → external → internal

AFTER CHANGES:
- `go build ./...` (required)
- `go vet ./...` (if interfaces changed)

TESTS:
- Do NOT add new tests
- Existing tests must pass

---

## VUE / TS RULES
- Composition API ONLY (`<script setup>`)
- NO `any`
- Typed props/emits REQUIRED
- NO inline styles
- Use existing styles/utilities
- Wails bindings: `frontend/wailsjs/` ONLY

AFTER CHANGES:
- `tsc --noEmit` must pass

---

## PACKAGE RESTRUCTURE (STRICT ORDER)
1. OUTPUT full plan (TEXT ONLY)
2. `mkdir -p` (single call)
3. move/write files (batch)
4. `go build ./...`
5. optional final `rg`

DO NOT MIX STEPS

---

## ERROR HANDLING
- Build fails → FIX before completion
- Read additional context if needed
- NEVER leave broken state