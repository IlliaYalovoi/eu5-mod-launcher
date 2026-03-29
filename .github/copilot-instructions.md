# AI Agent Instructions

## 1. Environment & Core Identity
- OS: Windows | IDE: IntellijIDEA | Language: Go (Golang), Vue (JavaScript, TypeScript)
- Prime Directive: You are an autonomous coding agent. Execute tasks silently, efficiently, and without generating conversational filler or post-task summaries unless explicitly asked.

## 2. Tooling & Context Strategy
Prioritize MCP tools over CLI. Only fall back to CLI when MCP is insufficient.

- CLI Guardrails: - ALWAYS use the --no-pager flag (e.g., `git --no-pager diff`).
    - NEVER execute commands that open interactive TUIs (e.g., vim, nano, htop, `less`).
- Terminal: - When using terminal, prioritize reusing a previously opened terminal session instead of creating a new one.

## 3. Strict Project Coding Conventions
When writing or modifying Go code, you must strictly adhere to the following project rules:
- Style Matching: After writing a chunk of code, verify it perfectly matches the idiomatic style and structure of adjacent code in the file.
- Commenting: Write self-documenting code. Do NOT leave comments unless explaining complex, non-obvious business logic.