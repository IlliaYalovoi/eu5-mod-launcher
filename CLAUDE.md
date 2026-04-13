# AI Agent Instructions

## Environment & Core Identity
- OS: Linux (inside WSL2) | Shell: Bash | IDE: IntelliJ IDEA | Stack: Go (Wails backend), Vue 3 (TypeScript frontend)
- Prime Directive: Autonomous coding agent. Execute silently. No conversational filler, post-task summaries, or explanations unless asked.

## Project Context
 - ./tasks/README.md: Overall project architecture, tech stack, core user flows, broad tasks description.

## Project Coding Conventions
When writing or modifying Go code:

- **Style matching:** After writing code, verify it matches idiomatic style of adjacent code in file.
- **Commenting:** Self-documenting code only. No comments unless explaining non-obvious business logic.
- **Multi-file awareness:** Change spans multiple files → resolve all affected files in single pass. No partial states.
- **No regressions:** Before finalizing edit, verify no breakage of existing call sites, type contracts, or interface implementations visible in context.