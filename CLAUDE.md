# AI Agent Instructions

## Environment
- OS: Linux (WSL2) | Shell: Bash | Stack: Go/Wails backend, Vue 3 + TypeScript frontend
- IDE: IntelliJ IDEA (do not generate IDE config files)

## Execution behavior
- No preamble, no post-task summaries, no progress narration
- Don't explain what you're about to do — just do it
- Don't ask clarifying questions mid-task unless blocked on something genuinely ambiguous
- On task completion: output only what changed and any unresolved issues
- DO NOT use worktrees, even if skills are suggesting it — just edit in place on main branch
- On failure: state what failed, why, and what you need — don't retry blindly more than twice

## Project context
- ./tasks/README.md: architecture, tech stack, user flows, task descriptions
- Read this file at session start if context about project goals is needed

## File reading
- Files ≤100 lines: read in full once
- Files >100 lines: read lines 1-50 first for structure, then targeted offset+limit reads
- Exception: if you need to rewrite a file completely, read it fully first regardless of size
- Never re-read a file already in context this session unless you have reason to believe it changed
- Go files: read only the relevant function unless doing a full-file rewrite
- Vue SFCs: read script/template/style sections separately on large files

## Shell tools
- ripgrep (`rg`) always, never `grep`
- `fd` always, never `find`
- Symbol search: `rg -n "Symbol" --type go`
- File discovery: `fd -e go`, `fd -e vue`
- Prefer CC native Read/Glob/Edit over Bash when the native tool covers it
- Bash only for what native tools cannot do

## Tool call discipline
- Batch all `mkdir` into one `mkdir -p path1 path2 path3` call
- Never `ls` to verify after write/mkdir — trust it succeeded
- Never `sed` for code edits — use Edit tool
- Never grep-then-sed — use Edit with exact old/new strings
- Write complete files in one Write call
- Plan all file operations first, then execute in batch
- One verification pass at the end if needed, never after each step
- Never use `ls` or `ls -la` to check if files/directories exist before operating on them
- Never use `ls` to discover what files are in a directory — use Glob tool instead
- `ls` is banned entirely. Use Glob for directory contents, Read for file existence.

## BANNED COMMANDS — never use these under any circumstances
- `ls` / `ls -la` / `ls -la path/` → use Glob tool
- `grep` → use `rg`
- `find` → use `fd`
- `sed -i` for code edits → use Edit tool
- Re-reading a file already in context → don't

Violation of these rules is a task failure.

## Parallel tool execution
- When multiple files need to be read independently, issue all Read calls in parallel in one turn
- When writing multiple independent files, issue all Write calls in parallel
- Never sequence tool calls that have no dependency on each other's output

## Go conventions
- Match idiomatic style of adjacent code in the same file
- Self-documenting names only — comments only for non-obvious business logic
- Changes spanning multiple files: resolve all affected files in one pass, no partial states
- Verify no broken call sites, type contracts, or interface implementations before finishing
- Import management: check existing imports before adding; use goimports conventions (stdlib → external → internal)
- After any non-trivial change: run `go build ./...` to verify compilation
- After interface changes: run `go vet ./...`
- No new tests unless explicitly requested; existing tests must still pass

## Vue/TypeScript conventions
- Composition API only (`<script setup>`) — no Options API
- Props and emits must be typed explicitly, no `any`
- No inline styles — use scoped CSS or existing utility classes
- Wails bindings are in `frontend/wailsjs/` — import from there, never reimplement
- After frontend changes: verify TypseScript compiles (`tsc --noEmit`)

## Package restructure pattern
1. Output full plan as text before touching any files
2. Create all directories in one `mkdir -p` call
3. Move/write all files
4. Run `go build ./...` to verify
5. One final `rg` pass only if needed
   Never interleave planning and execution.

## Error handling
- Build fails after your change: fix it before considering the task done
- If a fix requires understanding more context: read what you need, fix, verify
- Don't leave the codebase in a broken state between subtasks