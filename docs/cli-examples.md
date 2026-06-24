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

### Health snapshot export

```bash
todo-report health \
  --repo ~/lab/cswg/coordination \
  --branch jj \
  --compare main \
  --write-json coordination-health.json \
  --format text
```

This is useful when the team wants a checked-in JSON artifact to diff later
without giving up the normal terminal summary.

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

Limit discovery to one family of indexes:

```bash
todo-report indexes \
  --repo ~/lab/wire-lab \
  --branch main \
  --include-index SIM-beta \
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

Compare all discovered indexes across two branches:

```bash
todo-report health \
  --repo ~/lab/cswg/coordination \
  --branch jj \
  --all-indexes \
  --compare main \
  --format text
```

This is useful when the repo has multiple authoritative TODO roots and the team
needs repo-wide drift totals plus branch-only index visibility.

## Fleet example

Create a repo list file:

```text
~/lab/cswg/coordination
~/lab/wire-lab
```

Run a fleet-wide summary:

```bash
todo-report fleet health \
  --repo-list repos.txt \
  --branch main \
  --all-indexes \
  --format text
```

This is useful when the team wants one report across a portfolio of repos
instead of running the tool manually per repo.

Ad hoc fleet summary with inline repos:

```bash
todo-report fleet health \
  --repo ~/lab/cswg/coordination \
  --repo ~/lab/wire-lab \
  --branch main \
  --all-indexes \
  --format text
```

This is useful when a team wants a quick cross-repo report without creating or
editing a repo list file.

Fleet compare:

```bash
todo-report fleet health \
  --repo-list repos.txt \
  --branch main \
  --all-indexes \
  --compare jj \
  --format text
```

This is useful when a team wants repo-by-repo drift counts across a set of
working branches.

Fleet run with repo and index filters:

```bash
todo-report fleet health \
  --repo-list repos.txt \
  --branch main \
  --all-indexes \
  --include-repo wire-lab \
  --exclude-repo archive \
  --include-index protocols/ \
  --exclude-index retired/ \
  --format text
```

This is useful when the repo list is broad but the team only wants active repos
and active TODO trees in the current report.

Portable fleet Markdown:

```bash
todo-report fleet health \
  --repo-list repos.txt \
  --branch main \
  --all-indexes \
  --format markdown
```

This is useful when the team wants one report that reads cleanly in older Gitea
instances, GitHub, or plain Markdown viewers.

Exit-code oriented shell use:

```bash
todo-report drift \
  --repo ~/lab/cswg/coordination \
  --branch-a main \
  --branch-b jj \
  --format text
echo $?
```

This is useful when a script needs to distinguish "no branch drift" from
"differences found" without parsing the text output.

Fleet snapshot export:

```bash
todo-report fleet health \
  --repo-list repos.txt \
  --branch main \
  --all-indexes \
  --write-json fleet-health.json \
  --format text
```

This is useful when the team wants a stable JSON snapshot for Git history,
release notes, or later trend analysis.

## Known limitations

- Root-level `TODO/TODO.md` is still the default, but nested indexes are
  supported through `--index`.
- `health --all-indexes --compare` keeps open/completed/lint totals anchored to
  the selected branch and adds repo-wide drift and branch-only index lists for
  the comparison branch.
- `fleet health` currently supports repo lists through `--repo-list`; it does
  also supports repeated inline `--repo` flags, and the two inputs can be mixed.
- Repo and index filters are substring-based and case-insensitive.
- If filters remove every discovered index in an `--all-indexes` run, the
  report stays valid and returns zero discovered indexes.
- Exit codes are `0` for clean, `1` for warnings or differences, and `2` for
  errors or command failures.
- Cross-TODO references like `TODO-foo/bar` are not treated as missing subtasks
  in the current parent detail file.
- Checkbox-style detail subtasks accept `[ ]`, `[x]`, and `[~]`, with `[~]`
  reported as open.
