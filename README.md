# todo-report

`todo-report` is a small Go CLI for inspecting TODO health in local Git repos.
It is branch-aware, local-first, read-only, and designed as a pipes-and-filters
tool with human-readable output by default.

## What it does

The tool has three main features:

1. `age`
   Show how old top-level TODOs are, based on the first commit where each TODO
   appears in the selected branch's reachable history.
2. `drift`
   Compare TODO state across two branches, including top-level items, detail
   file presence, and subtasks.
3. `lint`
   Validate the TODO structure itself so broken links, malformed IDs, orphaned
   detail files, and similar issues are visible early.

`health` is the summary command on top of those three. It combines age, lint,
and optional branch drift into one report.

For monorepos, `indexes` discovers every authoritative `TODO/TODO.md` in the
selected branch, and `health --all-indexes` rolls them into a repo-wide summary.

## PromiseGrid relationship

`todo-report` is PromiseGrid-adjacent, not PromiseGrid-native.

It does not:

- create promises
- assess promises
- emit signed grid artifacts
- use CAS
- use `pCID`s

What it does share with the broader PromiseGrid style is:

- local-first observation
- explicit evidence from durable records
- branch and context awareness
- machine-readable output for later tooling

The intended integration path is indirect: `todo-report` emits local reports,
and later tools can import those reports as evidence.

## Repo assumptions

By default, `todo-report` assumes the inspected repo stores its active TODO
index at `TODO/TODO.md` relative to the Git repo root.

That means it works well for repos shaped like:

```text
repo/
  TODO/
    TODO.md
    TODO-binap-something.md
```

For repos whose active TODO index lives deeper in the tree, pass `--index` with
the authoritative index path relative to repo root, for example:

```bash
todo-report health --repo ~/lab/wire-lab --branch main --index protocols/wire-lab.d/TODO/TODO.md
```

Relative detail links inside that index are resolved from the index file's own
directory.

## Supported TODO styles

Top-level TODO IDs:

- proquint: `TODO-binap`
- numeric legacy: `001`, `1223`
- single-letter legacy prefix: `S122`

Detail-file subtask IDs keep their native form, including nested forms:

- `binap.1`
- `binap.2.1`
- `binap.2.1.1`
- `2.1`
- `Q-22.1`
- `UT-PSTK-origin`

Subtask hierarchy is not capped by the parser.
Checkbox-style detail subtasks accept `[ ]`, `[x]`, and `[~]`; `[~]` is treated
as open/in-progress for reporting purposes.

## Commands

### `age`

```bash
todo-report age --repo /path/to/repo --branch main
```

Shows top-level TODOs ordered by age, oldest first.
This currently includes both open and completed top-level TODOs.

### `drift`

```bash
todo-report drift --repo /path/to/repo --branch-a main --branch-b jj
```

Compares TODO state across two branches.

### `indexes`

```bash
todo-report indexes --repo /path/to/repo --branch main
```

Discovers root and nested `TODO/TODO.md` indexes on the selected branch.

### `lint`

```bash
todo-report lint --repo /path/to/repo --branch main
```

Validates the TODO structure rooted at the selected index file.

### `health`

```bash
todo-report health --repo /path/to/repo --branch main
todo-report health --repo /path/to/repo --branch main --compare jj
todo-report health --repo /path/to/repo --branch main --all-indexes
todo-report health --repo /path/to/repo --branch main --all-indexes --compare jj
```

Summarizes repo TODO health and can optionally include branch drift counts.
With `--all-indexes`, it discovers every authoritative TODO index on the
selected branch and reports a combined monorepo summary. With
`--all-indexes --compare`, totals stay anchored to the selected branch while
repo-wide drift and branch-only index lists are added for the comparison.

## Output formats

All commands support:

- `--format text`
- `--format markdown`
- `--format json`
- `--format tsv`

`--json` is also supported as a shortcut for `--format json`.

Format guidance:

- `text`: readable terminal output
- `markdown`: GitHub-friendly checklist and summary output
- `json`: structured export for later tooling
- `tsv`: Unix-friendly tab-separated output for shell pipelines

## Lint behavior

`lint` treats `TODO/TODO.md` as the source-of-truth index and also:

- validates referenced detail files
- reports duplicate TODO IDs
- reports duplicate subtask IDs
- reports malformed TODO IDs
- reports malformed subtask IDs on checkbox-style detail-file subtask lines
- reports invalid checkbox syntax
- reports missing detail files
- reports referenced subtasks not found within the same parent TODO detail file
- reports checked parents with open subtasks
- reports orphaned `TODO/TODO-*.md` detail files as warnings

## Examples

Human-readable summary:

```bash
todo-report health --repo ~/lab/cswg/coordination --branch jj
```

Multi-index monorepo summary:

```bash
todo-report health --repo ~/lab/wire-lab --branch main --all-indexes
```

Multi-index monorepo comparison:

```bash
todo-report health --repo ~/lab/wire-lab --branch main --all-indexes --compare jj
```

Markdown report for GitHub:

```bash
todo-report lint --repo ~/lab/cswg/coordination --branch jj --format markdown
```

Structured export:

```bash
todo-report health --repo ~/lab/cswg/coordination --branch jj --format json
```

TSV for shell tools:

```bash
todo-report age --repo ~/lab/cswg/coordination --branch jj --format tsv
```

Index discovery:

```bash
todo-report indexes --repo ~/lab/wire-lab --branch main --format text
```

Nested-index repo:

```bash
todo-report health --repo ~/lab/wire-lab --branch main --index protocols/wire-lab.d/TODO/TODO.md
```

More concrete examples live in [docs/cli-examples.md](docs/cli-examples.md).

## Demo video

A rendered landscape demo video is checked in at
[docs/videos/todo-report-demo.mp4](docs/videos/todo-report-demo.mp4).

The generator script lives at
[scripts/render-demo-video.sh](scripts/render-demo-video.sh).
It renders a terminal-style explainer video using the `coordination` repo as
the demo source.

## Development

Run tests with:

```bash
go test ./...
```

Run coverage with:

```bash
go test ./... -cover
```
