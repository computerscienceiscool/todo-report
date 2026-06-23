# todo-report

`todo-report` is a small Go CLI for inspecting TODO health in local Git repos.
It is branch-aware, local-first, and read-only.

## Commands

- `todo-report age --repo /path/to/repo --branch main`
- `todo-report drift --repo /path/to/repo --branch-a main --branch-b jj`
- `todo-report lint --repo /path/to/repo --branch main`
- `todo-report health --repo /path/to/repo --branch main --compare jj`

## Output formats

All commands support:

- `--format text`
- `--format markdown`
- `--format json`
- `--format tsv`

`--json` is also supported as a shortcut for `--format json`.

## Supported TODO styles

Top-level TODO IDs:

- proquint style: `TODO-binap`
- numeric legacy style: `001`, `1223`
- single-letter legacy prefix: `S122`

Detail-file subtasks keep their native IDs, including nested forms:

- `binap.1`
- `binap.2.1`
- `2.1`

## Lint behavior

`lint` treats `TODO/TODO.md` as the source-of-truth index and also:

- validates referenced detail files
- reports duplicate TODO IDs and duplicate subtask IDs
- reports malformed IDs and invalid checkboxes
- reports checked parents with open subtasks
- reports orphaned `TODO/TODO-*.md` detail files as warnings

## Development

Run tests with:

```bash
go test ./...
```
