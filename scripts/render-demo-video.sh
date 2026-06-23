#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
target_repo="${1:-$HOME/lab/cswg/coordination}"
branch="${2:-$(git -C "$target_repo" branch --show-current)}"

work_dir="$repo_root/docs/demo-video"
slides_dir="$work_dir/slides"
renders_dir="$work_dir/renders"
output_file="$repo_root/docs/videos/todo-report-demo.mp4"

mkdir -p "$slides_dir" "$renders_dir" "$(dirname "$output_file")"
rm -f "$slides_dir"/*.txt "$renders_dir"/*.mp4 "$renders_dir"/concat.txt

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
  (
    cd "$repo_root"
    go run ./cmd/todo-report "$subcommand" "$@"
  )
}

cat >"$slides_dir/00-intro.txt" <<EOF
todo-report

Demo repo: $target_repo
Branch: $branch

This short video shows the tool's three big features:
1. age
2. drift
3. lint

It also shows the health summary and one structured export.
EOF

age_output="$(run_report age --repo "$target_repo" --branch "$branch" --format text | sed -n '1,12p')"
cat >"$slides_dir/10-age.txt" <<EOF
AGE REPORT

Feature 1: find old open work from Git history.
This is branch-local and only covers top-level TODOs.

\$ todo-report age --repo $target_repo --branch $branch --format text

$age_output
EOF

drift_output="$(run_report drift --repo "$target_repo" --branch-a main --branch-b "$branch" --format text | sed -n '1,28p')"
cat >"$slides_dir/20-drift.txt" <<EOF
DRIFT REPORT

Feature 2: compare TODO state across branches.
This shows branch-specific TODOs and subtask divergence.

\$ todo-report drift --repo $target_repo --branch-a main --branch-b $branch --format text

$drift_output
EOF

lint_output="$(run_report lint --repo "$target_repo" --branch "$branch" --format text)"
cat >"$slides_dir/30-lint.txt" <<EOF
LINT REPORT

Feature 3: validate the TODO structure itself.
This catches broken links and orphaned detail files.

\$ todo-report lint --repo $target_repo --branch $branch --format text

$lint_output
EOF

health_output="$(run_report health --repo "$target_repo" --branch "$branch" --format json | sed -n '1,36p')"
cat >"$slides_dir/40-health.txt" <<EOF
HEALTH REPORT + JSON EXPORT

Health combines age, lint, and optional drift into one summary.
This also shows machine-readable output for scripts and later tooling.

\$ todo-report health --repo $target_repo --branch $branch --format json

$health_output
EOF

render_slide() {
  local text_file="$1"
  local duration="$2"
  local out_file="$3"

  ffmpeg -y \
    -f lavfi -i "color=c=${bg_color}:s=${video_size}:d=${duration}" \
    -vf "drawbox=x=30:y=28:w=1220:h=664:color=${panel_color}:t=fill,\
drawbox=x=48:y=170:w=1184:h=494:color=${terminal_color}:t=fill,\
drawtext=fontfile=${font_file}:text='todo-report':fontcolor=${brand_color}:fontsize=30:x=56:y=44,\
drawtext=fontfile=${font_file}:text='coordination repo demo':fontcolor=${accent_color}:fontsize=18:x=1080:y=46,\
drawtext=fontfile=${font_file}:textfile=${text_file}:fontcolor=${text_color}:fontsize=22:line_spacing=10:x=64:y=92" \
    -c:v libx264 -pix_fmt yuv420p "$out_file" >/dev/null 2>&1
}

render_slide "$slides_dir/00-intro.txt" 4 "$renders_dir/00-intro.mp4"
render_slide "$slides_dir/10-age.txt" 6 "$renders_dir/10-age.mp4"
render_slide "$slides_dir/20-drift.txt" 7 "$renders_dir/20-drift.mp4"
render_slide "$slides_dir/30-lint.txt" 6 "$renders_dir/30-lint.mp4"
render_slide "$slides_dir/40-health.txt" 8 "$renders_dir/40-health.mp4"

cat >"$renders_dir/concat.txt" <<EOF
file '$renders_dir/00-intro.mp4'
file '$renders_dir/10-age.mp4'
file '$renders_dir/20-drift.mp4'
file '$renders_dir/30-lint.mp4'
file '$renders_dir/40-health.mp4'
EOF

ffmpeg -y -f concat -safe 0 -i "$renders_dir/concat.txt" -c copy "$output_file" >/dev/null 2>&1

printf 'Wrote %s\n' "$output_file"
