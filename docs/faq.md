# FAQ

## What problem does `todo-report` solve?

It gives a team one CLI for answering four practical questions:

- how old are our TODOs?
- how do TODOs differ across branches?
- is the TODO structure valid?
- what is the overall TODO health of this repo, monorepo, or fleet?

## Does it modify the repo?

No. It is read-only.

It reads:

- Git history
- the selected TODO index
- linked detail files

It can optionally write a JSON snapshot file when you pass `--write-json`.

## What is the source of truth?

The selected `TODO/TODO.md` index is the source of truth for what the active
TODO set is.

Detail files add subtask and status evidence, but they do not replace the role
of the index.

## Why does `health` care about branches?

Because TODO state can legitimately differ across branches.

A top-level item can be complete in one branch and still open in another. The
tool treats branch context as real data, not noise.

## Why is `health --all-indexes` useful?

It is the monorepo mode.

Use it when a repo contains multiple authoritative TODO roots and you want one
combined report without losing the per-index breakdown.

## Why is `fleet health` useful?

It is the multi-repo mode.

Use it when the team wants a portfolio view across many repos, either from a
`repos.txt` file or repeated `--repo` flags.

## What does `detect` do?

`detect` is a compatibility scanner.

It reports:

- top-level ID styles
- subtask styles
- index layout
- compatibility findings

Use it before adopting the tool in a new repo or when you suspect the repo uses
an older or mixed TODO dialect.

## What do the exit codes mean?

- `0`: clean
- `1`: warnings or differences
- `2`: errors or command failure

This makes the tool usable in shell scripts, cron jobs, and CI.

## Does `lint` catch index/detail drift?

Yes.

It reports:

- `index_open_detail_complete`
- `index_done_detail_open`

Those are especially useful when people or LLMs update the index and detail
files separately.

## Does it work only with root-level `TODO/TODO.md`?

No.

That is the default, but nested indexes are supported through `--index`, and
monorepos can be scanned with `indexes` and `health --all-indexes`.
