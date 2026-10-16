# Skills installed by the AI Software Factory

Copied unchanged (minus authoring test logs and one TypeScript example file) into
every product's `.claude/skills/` so that whoever works on the product in Claude
Code keeps the same code discipline the factory's agents are held to (the
"Code discipline" section of `PROJECT_RULES.md`). The factory's own agents get the
ones `agents/addons.toml` names, injected into their prompts; a skill added there
needs its licence here first (the eval `addons-skills-licensed` checks the entry).

| Skill | What it prevents | Source | License |
|---|---|---|---|
| ponytail | code that need not exist: YAGNI ladder, reuse before writing, shortest working diff | ponytail 4.8.3 (DietrichGebert) | MIT, `licenses/ponytail.LICENSE` |
| ponytail-review | over-engineering in review: unrequested abstractions, dead options | ponytail 4.8.3 | MIT, `licenses/ponytail.LICENSE` |
| karpathy-guidelines | silent assumptions, overcomplication, drive-by edits, unverifiable goals | andrej-karpathy-skills 1.0.0 | MIT (skill frontmatter) |
| test-driven-development | code without a failing test first | superpowers 6.4.1 (Jesse Vincent) | MIT, `licenses/superpowers.LICENSE` |
| systematic-debugging | symptom patches instead of root-cause fixes | superpowers 6.4.1 | MIT, `licenses/superpowers.LICENSE` |
| verification-before-completion | "done" claims without evidence | superpowers 6.4.1 | MIT, `licenses/superpowers.LICENSE` |

## Written by the factory for its own agents

Not copied into products: these are injected into the agents `agents/addons.toml` names.
Their text is the factory's own; each one says what went wrong that it answers.

| Skill | What it prevents | Given to | Licence |
|---|---|---|---|
| reuse-before-copy | a helper, fixture or constant copied into each feature instead of shared | architect, coder | factory-authored |
| insecure-defaults | fallback secrets, default credentials, fail-open switches, weak crypto, permissive access, leaking and swallowed errors | boundary, coder, tester | factory-authored; the six categories follow Trail of Bits' `insecure-defaults` (CC BY-SA 4.0), the text does not copy it |
| test-strength | tests that would still pass when the behaviour breaks | tester | factory-authored |
| challenge-speculation | a "maybe later" built into every story as if it were a requirement | interview, spec | factory-authored |
