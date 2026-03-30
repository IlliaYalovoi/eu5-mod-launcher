# AI Agent Instructions

## 1. Environment & Core Identity
- OS: Windows | IDE: IntellijIDEA | Stack: Go (Wails backend), Vue 3 (TypeScript frontend)
- Prime Directive: You are an autonomous coding agent. Execute tasks silently and efficiently. Do NOT produce conversational filler, post-task summaries, or explanations unless explicitly requested.

## 2. Tooling & Context Strategy
Prioritize MCP tools over CLI. Fall back to CLI only when MCP is insufficient.

- CLI Guardrails:
  - ALWAYS use the `--no-pager` flag (e.g., `git --no-pager diff`).
  - NEVER execute commands that open interactive TUIs (e.g., vim, nano, htop, less).
- Terminal: Reuse a previously opened terminal session instead of creating a new one.

## 3. Strict Project Coding Conventions
When writing or modifying Go code, strictly adhere to the following rules:

- Style Matching: After writing any code, verify it perfectly matches the idiomatic style and structure of adjacent code in the file.
- Commenting: Write self-documenting code. Do NOT add comments unless explaining complex, non-obvious business logic.
- Multi-file awareness: When a change spans multiple files (e.g., Wails bindings + Vue frontend), resolve all affected files in a single pass. Do not leave partial states.
- No regressions: Before finalizing any edit, verify it does not break existing call sites, type contracts, or interface implementations visible in context.