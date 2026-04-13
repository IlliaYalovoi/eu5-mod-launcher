## CORE BEHAVIOR
- Using `caveman` skill is MANDATORY for all tasks
- ALWAYS use task appropriate skills/superpowers
- NO finishing task early if issues remain
- OUTPUT ONLY:
    - file changes
    - unresolved issues (if any)
- TERMINATE immediately after finishing task

## ENVIRONMENT
- OS: Linux (WSL2) | Shell: Bash
- Stack: Go (Wails), Vue 3 + TypeScript
- Test environment: Windows 11 with EU5, EU4, HOI4, Vic3, CK3, Stellaris installed (No Linux/WSL2/MacOS runs on active development stage)

## PROJECT
- Desktop app for managing mods across multiple Paradox games titles.
- For legacy sqlite games must use `launcher-v2.sqlite` as db NOT `launcher-v2.db`.

## EXECUTION RULES
- Edit in-place on main branch (NO worktrees)
- Plan internally, execute in batches
- DO NOT interleave planning and execution
- On failure:
    - state failure
    - state cause
    - state requirement

## TOOLING RULES
- Prefer native tools (Read/Glob/Edit) over Bash
- Bash only when necessary

## GO RULES
- Match local code style
- Self-documenting names (minimal comments)
- No partial states across files

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
