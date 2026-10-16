# Jsonschema 2020 12 Project Rules

- Project ID: PROJ-010
- Slug: jsonschema-2020-12
- Source root: the repository root (write repo-relative paths, e.g. `app/main.py`)
- Factory-owned evidence (never write these): `docs/work/`, `docs/architecture/adr/`, `docs/releases/`, `PROJECT_RULES.md`, `project-spec.json`, `.claude/skills/`

## Code discipline

- Build only what the story's acceptance criteria require: no speculative features,
  options, or scaffolding "for later".
- Reuse before writing: this repo's code, then the standard library, then an installed
  dependency; add a dependency only when a few lines cannot do it.
- No abstraction with one implementation, no config for a value that never changes.
- Surgical changes: touch only what the task needs, match the surrounding style, never
  refactor or reformat unrelated code.
- Fix a bug at its root, where every caller benefits, and start with a test that
  reproduces it. Every behaviour change ships with the smallest test that fails if it
  breaks.
- A feature's tests go in that feature's own test file: start a new file rather than
  growing a general one that every later change must then touch.
- A test asks the code for a value the code defines (a size, a limit, a default)
  instead of repeating the number, so changing it once does not break unrelated tests.
- Never trade away input validation at trust boundaries, error handling that prevents
  data loss, or security measures for brevity.
- Claim only what was verified.

(The full practice: `.claude/skills/` — ponytail, karpathy-guidelines, superpowers TDD,
systematic debugging, verification before completion.)
