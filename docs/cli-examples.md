# CLI Examples

This file records real-world example invocations for `todo-report`.

## Coordination repo

The current demo repo is `~/lab/cswg/coordination` on branch `jj`.

### Age

```bash
todo-report age --repo ~/lab/cswg/coordination --branch jj --format text
```

This is useful when the team wants to see stale open work ordered by age.

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
broken detail links or orphaned detail files.

### Health JSON

```bash
todo-report health --repo ~/lab/cswg/coordination --branch jj --format json
```

This is useful when another tool wants a structured export with summary counts,
lint findings, and age data in one payload.

## Known limitation

The current implementation expects `TODO/TODO.md` at repo root. Repos such as
`wire-lab` that keep active TODO indexes deeper in the tree are not yet
supported by the current CLI contract.
