# CLI Examples

This file records real-world example invocations for `todo-report`.

Unless otherwise noted, `age` currently reports all top-level TODOs found in
the selected branch, not only open ones.

## Coordination repo

The current demo repo is `~/lab/cswg/coordination` on branch `jj`.

### Age

```bash
todo-report age --repo ~/lab/cswg/coordination --branch jj --format text
```

This is useful when the team wants to see older top-level TODOs ordered by age.

### Drift

```bash
todo-report drift --repo ~/lab/cswg/coordination --branch-a main --branch-b jj --format text
```

This is useful when one branch still uses older numeric TODOs and another has
moved to proquint TODOs or otherwise diverged in structure.

### Lint

```bash
todo-report lint --repo ~/lab/cswg/coordination --branch jj --format markdown
```

This is useful when the team wants a GitHub-friendly checklist-style report of
broken detail links, malformed subtask lines, or orphaned detail files.

### Health JSON

```bash
todo-report health --repo ~/lab/cswg/coordination --branch jj --format json
```

This is useful when another tool wants a structured export with summary counts,
lint findings, and age data in one payload.

### Health text

```bash
todo-report health --repo ~/lab/cswg/coordination --branch jj --format text
```

This is useful when the team wants a concise narrative summary with status,
top finding types, and the oldest open TODOs in one screenful.

## Nested index example

`wire-lab` keeps its active master queue under a nested path, so it needs
`--index`:

```bash
todo-report health \
  --repo ~/lab/wire-lab \
  --branch main \
  --index protocols/wire-lab.d/TODO/TODO.md \
  --format text
```

This is useful for monorepos where the authoritative TODO index is not at
repo-root `TODO/TODO.md`.

## Multi-index monorepo example

Discover every authoritative TODO index in `wire-lab`:

```bash
todo-report indexes \
  --repo ~/lab/wire-lab \
  --branch main \
  --format text
```

Summarize all discovered indexes together:

```bash
todo-report health \
  --repo ~/lab/wire-lab \
  --branch main \
  --all-indexes \
  --format text
```

This is useful for monorepos that have one master queue plus multiple nested or
simulation-specific TODO roots.

## Known limitations

- Root-level `TODO/TODO.md` is still the default, but nested indexes are
  supported through `--index`.
- `health --all-indexes` discovers indexes on the selected branch and does not
  yet combine that mode with `--compare`.
- Cross-TODO references like `TODO-foo/bar` are not treated as missing subtasks
  in the current parent detail file.
- Checkbox-style detail subtasks accept `[ ]`, `[x]`, and `[~]`, with `[~]`
  reported as open.
