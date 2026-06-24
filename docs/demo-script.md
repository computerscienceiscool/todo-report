# Demo Script

This is a copy-paste demo flow for showing `todo-report` to a team.

Primary demo repos:

- single repo: `~/lab/cswg/coordination`
- monorepo: `~/lab/wire-lab`

## 1. One-sentence intro

`todo-report` is a branch-aware CLI that measures TODO age, compares drift,
lints TODO structure, and rolls those signals into a health summary.

## 2. Show the four core views

### Age

```bash
todo-report age --repo ~/lab/cswg/coordination --branch jj --format text
```

Talk track:

- shows oldest top-level TODOs first
- useful for spotting stale work

### Drift

```bash
todo-report drift --repo ~/lab/cswg/coordination --branch-a main --branch-b jj --format text
```

Talk track:

- compares TODO state across branches
- branch context matters; completion is not global

### Lint

```bash
todo-report lint --repo ~/lab/cswg/coordination --branch jj --format markdown
```

Talk track:

- catches broken links, orphan detail files, malformed IDs
- also catches index/detail completion mismatches

### Health

```bash
todo-report health --repo ~/lab/cswg/coordination --branch jj --format text
```

Talk track:

- this is the easiest “daily driver” command
- gives one summary instead of forcing the team to run age, drift, and lint separately

## 3. Show structured output

```bash
todo-report health --repo ~/lab/cswg/coordination --branch jj --format json
```

Talk track:

- same information, machine-readable
- easy to export into other tooling

## 4. Show monorepo support

### Discover indexes

```bash
todo-report indexes --repo ~/lab/wire-lab --branch main --format text
```

### Summarize all indexes

```bash
todo-report health --repo ~/lab/wire-lab --branch main --all-indexes --format text
```

Talk track:

- this is how the tool handles repos with more than one authoritative TODO root

## 5. Show fleet mode

### File-based fleet

```bash
todo-report fleet health --repo-list repos.txt --branch main --all-indexes --format markdown
```

### Ad hoc fleet

```bash
todo-report fleet health \
  --repo ~/lab/cswg/coordination \
  --repo ~/lab/wire-lab \
  --branch main \
  --all-indexes \
  --format text
```

Talk track:

- `--repo-list` is good for repeatable team runs
- repeated `--repo` is good for ad hoc use during a meeting

## 6. Show compatibility detection

```bash
todo-report detect --repo ~/lab/cswg/coordination --branch jj --format text
```

Talk track:

- use this when onboarding a new repo
- it tells you what TODO dialect the repo is using

## 7. Show filters

```bash
todo-report fleet health \
  --repo-list repos.txt \
  --branch main \
  --all-indexes \
  --include-repo wire-lab \
  --exclude-index retired/ \
  --format text
```

Talk track:

- useful when the repo list is broad but you only want active work

## 8. Show snapshot export

```bash
todo-report fleet health \
  --repo-list repos.txt \
  --branch main \
  --all-indexes \
  --write-json fleet-health.json \
  --format text
```

Talk track:

- saves a stable JSON artifact without changing normal stdout output

## 9. Show exit codes

```bash
todo-report drift --repo ~/lab/cswg/coordination --branch-a main --branch-b jj --format text
echo $?
```

Talk track:

- `0` clean
- `1` warnings or differences
- `2` errors

## 10. Short close

Recommended close:

`todo-report` works at three levels: one repo, one monorepo, or a whole fleet,
and it stays useful for both humans and scripts.
