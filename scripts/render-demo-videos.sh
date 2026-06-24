#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

coord_repo="${1:-$HOME/lab/cswg/coordination}"
coord_branch="${2:-jj}"
wire_repo="${3:-$HOME/lab/wire-lab}"
wire_branch="${4:-main}"

work_dir="$repo_root/docs/demo-videos"
slides_dir="$work_dir/slides"
renders_dir="$work_dir/renders"
videos_dir="$repo_root/docs/videos"

mkdir -p "$slides_dir" "$renders_dir" "$videos_dir"
rm -rf "$slides_dir" "$renders_dir"
mkdir -p "$slides_dir" "$renders_dir"

font_file="$(fc-match monospace -f '%{file}\n' | head -n 1)"
bg_color="0x0d1117"
panel_color="0x161b22"
terminal_color="0x010409"
brand_color="0x58a6ff"
text_color="0xc9d1d9"
title_color="0xf0f6fc"
accent_color="0x7ee787"
video_size="1280x720"

run_report() {
  local subcommand="$1"
  shift
  local output
  set +e
  output="$(
    cd "$repo_root" &&
      go run ./cmd/todo-report "$subcommand" "$@" 2>/dev/null
  )"
  set -e
  printf '%s\n' "$output"
}

render_slide() {
  local text_file="$1"
  local duration="$2"
  local out_file="$3"
  local label="$4"

  ffmpeg -y \
    -f lavfi -i "color=c=${bg_color}:s=${video_size}:d=${duration}" \
    -vf "drawbox=x=30:y=28:w=1220:h=664:color=${panel_color}:t=fill,\
drawbox=x=48:y=170:w=1184:h=494:color=${terminal_color}:t=fill,\
drawtext=fontfile=${font_file}:text='todo-report':fontcolor=${brand_color}:fontsize=30:x=56:y=44,\
drawtext=fontfile=${font_file}:text='${label}':fontcolor=${accent_color}:fontsize=18:x=990:y=46,\
drawtext=fontfile=${font_file}:textfile=${text_file}:fontcolor=${text_color}:fontsize=22:line_spacing=10:x=64:y=92" \
    -c:v libx264 -pix_fmt yuv420p "$out_file" >/dev/null 2>&1
}

concat_video() {
  local list_file="$1"
  local output_file="$2"
  ffmpeg -y -f concat -safe 0 -i "$list_file" -c copy "$output_file" >/dev/null 2>&1
}

core_age="$(run_report age --repo "$coord_repo" --branch "$coord_branch" --format text | sed -n '1,8p')"
core_drift="$(run_report drift --repo "$coord_repo" --branch-a main --branch-b "$coord_branch" --format text | sed -n '1,18p')"
core_lint="$(run_report lint --repo "$coord_repo" --branch "$coord_branch" --format text | sed -n '1,12p')"
core_health_json="$(run_report health --repo "$coord_repo" --branch "$coord_branch" --format json | sed -n '1,22p')"

mkdir -p "$slides_dir/core" "$renders_dir/core"
cat >"$slides_dir/core/00-intro.txt" <<EOF
todo-report

Core demo

Repo: $coord_repo
Branch: $coord_branch

This walkthrough shows the four daily-driver views:
1. age
2. drift
3. lint
4. health

All slides are intentionally slowed down for live narration.
EOF

cat >"$slides_dir/core/10-age.txt" <<EOF
AGE

Find older top-level TODOs using Git history.
Useful for spotting stale work.

\$ todo-report age --repo $coord_repo --branch $coord_branch --format text

$core_age
EOF

cat >"$slides_dir/core/20-drift.txt" <<EOF
DRIFT

Compare TODO state across branches.
Branch context matters; completion is not global.

\$ todo-report drift --repo $coord_repo --branch-a main --branch-b $coord_branch --format text

$core_drift
EOF

cat >"$slides_dir/core/30-lint.txt" <<EOF
LINT

Validate structure and cross-file consistency.
This catches malformed IDs, broken links, orphan files,
and index/detail completion drift.

\$ todo-report lint --repo $coord_repo --branch $coord_branch --format text

$core_lint
EOF

cat >"$slides_dir/core/40-health.txt" <<EOF
HEALTH + JSON EXPORT

Health rolls age, lint, and optional drift into one report.
The same command can emit structured JSON for later tooling.

\$ todo-report health --repo $coord_repo --branch $coord_branch --format json

$core_health_json
EOF

render_slide "$slides_dir/core/00-intro.txt" 7 "$renders_dir/core/00-intro.mp4" "core demo"
render_slide "$slides_dir/core/10-age.txt" 9 "$renders_dir/core/10-age.mp4" "core demo"
render_slide "$slides_dir/core/20-drift.txt" 10 "$renders_dir/core/20-drift.mp4" "core demo"
render_slide "$slides_dir/core/30-lint.txt" 10 "$renders_dir/core/30-lint.mp4" "core demo"
render_slide "$slides_dir/core/40-health.txt" 11 "$renders_dir/core/40-health.mp4" "core demo"

cat >"$renders_dir/core/concat.txt" <<EOF
file '$renders_dir/core/00-intro.mp4'
file '$renders_dir/core/10-age.mp4'
file '$renders_dir/core/20-drift.mp4'
file '$renders_dir/core/30-lint.mp4'
file '$renders_dir/core/40-health.mp4'
EOF

concat_video "$renders_dir/core/concat.txt" "$videos_dir/todo-report-core-demo.mp4"

mono_indexes="$(run_report indexes --repo "$wire_repo" --branch "$wire_branch" --format text | sed -n '1,12p')"
mono_health="$(run_report health --repo "$wire_repo" --branch "$wire_branch" --all-indexes --exclude-index archive/ --format text | sed -n '1,16p')"
mono_fleet="$(run_report fleet health --repo "$coord_repo" --repo "$wire_repo" --branch "$wire_branch" --all-indexes --exclude-index archive/ --format text | sed -n '1,16p')"

mkdir -p "$slides_dir/monorepo" "$renders_dir/monorepo"
cat >"$slides_dir/monorepo/00-intro.txt" <<EOF
todo-report

Monorepo + fleet demo

Repos:
- $wire_repo
- $coord_repo

This walkthrough shows:
1. index discovery
2. repo-wide health across multiple TODO roots
3. fleet health across multiple repos
EOF

cat >"$slides_dir/monorepo/10-indexes.txt" <<EOF
INDEXES

Discover every authoritative TODO index in a monorepo.

\$ todo-report indexes --repo $wire_repo --branch $wire_branch --format text

$mono_indexes
EOF

cat >"$slides_dir/monorepo/20-health.txt" <<EOF
HEALTH --ALL-INDEXES

Summarize multiple TODO roots together.
This example excludes archived paths to keep the report focused.

\$ todo-report health --repo $wire_repo --branch $wire_branch --all-indexes --exclude-index archive/ --format text

$mono_health
EOF

cat >"$slides_dir/monorepo/30-fleet.txt" <<EOF
FLEET HEALTH

Roll multiple repos into one summary.
Inline --repo flags make ad hoc team demos easy.

\$ todo-report fleet health --repo $coord_repo --repo $wire_repo --branch $wire_branch --all-indexes --exclude-index archive/ --format text

$mono_fleet
EOF

render_slide "$slides_dir/monorepo/00-intro.txt" 7 "$renders_dir/monorepo/00-intro.mp4" "monorepo + fleet"
render_slide "$slides_dir/monorepo/10-indexes.txt" 9 "$renders_dir/monorepo/10-indexes.mp4" "monorepo + fleet"
render_slide "$slides_dir/monorepo/20-health.txt" 10 "$renders_dir/monorepo/20-health.mp4" "monorepo + fleet"
render_slide "$slides_dir/monorepo/30-fleet.txt" 10 "$renders_dir/monorepo/30-fleet.mp4" "monorepo + fleet"

cat >"$renders_dir/monorepo/concat.txt" <<EOF
file '$renders_dir/monorepo/00-intro.mp4'
file '$renders_dir/monorepo/10-indexes.mp4'
file '$renders_dir/monorepo/20-health.mp4'
file '$renders_dir/monorepo/30-fleet.mp4'
EOF

concat_video "$renders_dir/monorepo/concat.txt" "$videos_dir/todo-report-monorepo-fleet-demo.mp4"

compat_detect="$(run_report detect --repo "$wire_repo" --branch "$wire_branch" --index protocols/wire-lab.d/TODO/TODO.md --format text | sed -n '1,14p')"

mkdir -p "$slides_dir/compat" "$renders_dir/compat"
cat >"$slides_dir/compat/00-intro.txt" <<EOF
todo-report

Compatibility + adoption demo

Repo: $wire_repo
Index: protocols/wire-lab.d/TODO/TODO.md

This walkthrough shows how the tool helps teams adopt
mixed TODO dialects instead of assuming one perfect format.
EOF

cat >"$slides_dir/compat/10-detect.txt" <<EOF
DETECT

Inspect a repo's TODO dialect before tightening lint rules
or rolling the tool out across a fleet.

\$ todo-report detect --repo $wire_repo --branch $wire_branch --index protocols/wire-lab.d/TODO/TODO.md --format text

$compat_detect
EOF

cat >"$slides_dir/compat/20-styles.txt" <<EOF
SUPPORTED TOP-LEVEL STYLES

todo-report currently supports mixed real-world team styles:

- TODO-binap
- jirin
- 001
- S122
- 026-planning-group-workspace-mvp.md

This makes adoption easier across older and newer repos.
EOF

cat >"$slides_dir/compat/30-close.txt" <<EOF
WHY THIS MATTERS

Adoption does not start with one perfect repo.
It starts with:

- detect
- compatibility reporting
- linted index/detail consistency
- branch-aware health summaries

That is the path from one repo to a team-wide fleet.
EOF

render_slide "$slides_dir/compat/00-intro.txt" 7 "$renders_dir/compat/00-intro.mp4" "compatibility demo"
render_slide "$slides_dir/compat/10-detect.txt" 10 "$renders_dir/compat/10-detect.mp4" "compatibility demo"
render_slide "$slides_dir/compat/20-styles.txt" 9 "$renders_dir/compat/20-styles.mp4" "compatibility demo"
render_slide "$slides_dir/compat/30-close.txt" 8 "$renders_dir/compat/30-close.mp4" "compatibility demo"

cat >"$renders_dir/compat/concat.txt" <<EOF
file '$renders_dir/compat/00-intro.mp4'
file '$renders_dir/compat/10-detect.mp4'
file '$renders_dir/compat/20-styles.mp4'
file '$renders_dir/compat/30-close.mp4'
EOF

concat_video "$renders_dir/compat/concat.txt" "$videos_dir/todo-report-compat-demo.mp4"

printf 'Wrote %s\n' "$videos_dir/todo-report-core-demo.mp4"
printf 'Wrote %s\n' "$videos_dir/todo-report-monorepo-fleet-demo.mp4"
printf 'Wrote %s\n' "$videos_dir/todo-report-compat-demo.mp4"
