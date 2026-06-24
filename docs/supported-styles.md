# Supported TODO Styles

This document describes the TODO index and detail-file styles that
`todo-report` currently understands.

## Source of truth

The authoritative queue is always the selected index file:

- default: `TODO/TODO.md`
- nested repos: pass `--index path/to/TODO/TODO.md`

Linked detail files are read relative to the index file's directory.

## Top-level TODO IDs

Supported top-level ID styles:

- prefixed proquint: `TODO-binap`
- bare proquint: `jirin`
- numeric legacy: `001`, `1223`
- single-letter legacy prefix: `S122`
- filename stem: `026-planning-group-workspace-mvp.md`

Examples:

```markdown
- [ ] TODO-binap - Lock outline (`TODO/TODO-binap.md`)
- [ ] jirin - decomk bootstrap (`TODO/TODO-jirin.md`)
- [ ] 001 - Trial shared Codespaces access (`TODO/001-shared-codespaces.md`)
- [ ] S122 - Spelling check (`TODO/S122-spelling-check.md`)
- [ ] 026-planning-group-workspace-mvp.md Planning group workspace tool MVP
```

## Index layouts

Supported index layouts:

- checklist lines
- Markdown table rows with linked handles

Checklist example:

```markdown
- [ ] TODO-binap - Lock outline (`TODO/TODO-binap.md`)
```

Table example:

```markdown
| [TODO-hipak](./TODO-hipak-local.md) | 2026-05-25 | Local title | — |
```

## Detail-file subtasks

Supported checkbox states:

- `[ ]` open
- `[x]` complete
- `[~]` in progress, treated as open for reporting

Supported subtask ID styles:

- parent-prefixed dotted: `binap.1`, `binap.2.1`
- numeric dotted: `2.1`
- token or dotted token: `Q-22.1`, `UT-PSTK-origin`

Examples:

```markdown
- [ ] binap.1 First step
- [x] binap.2.1 Nested step
- [ ] 2.1 Legacy nested step
- [ ] UT-PSTK-origin Single-token step
- [~] 1. Approximate step
```

## Supported cross-file consistency checks

`lint` checks more than syntax.

It currently reports:

- `missing_detail_file`
- `orphan_detail_file`
- `duplicate_id`
- `duplicate_subtask_id`
- `malformed_todo_id`
- `malformed_subtask_id`
- `invalid_checkbox`
- `referenced_subtask_not_found`
- `index_open_detail_complete`
- `index_done_detail_open`

`index_open_detail_complete` means the index still shows the parent item as open
while the linked detail file appears complete.

`index_done_detail_open` means the index shows the parent item as done while the
linked detail file still contains open or in-progress subtasks.

## Current limits

- `age` currently reports all top-level TODOs, not only open ones.
- Cross-TODO references like `TODO-foo/bar` are not treated as missing local
  subtasks in the current detail file.
- The tool does not require one single style per repo. Mixed styles are allowed
  and are surfaced by `detect`.
