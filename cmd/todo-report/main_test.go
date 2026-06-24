package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"todo-report/internal/testrepo"
)

func TestRunErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "no args",
			args: nil,
			want: "usage:",
		},
		{
			name: "unknown command",
			args: []string{"bogus"},
			want: "usage:",
		},
		{
			name: "age missing branch",
			args: []string{"age"},
			want: "age requires --branch",
		},
		{
			name: "drift missing branches",
			args: []string{"drift"},
			want: "drift requires --branch-a and --branch-b",
		},
		{
			name: "lint unsupported format",
			args: []string{"lint", "--branch", "main", "--format", "xml"},
			want: `unsupported format "xml"`,
		},
		{
			name: "health missing branch",
			args: []string{"health"},
			want: "health requires --branch",
		},
		{
			name: "fleet missing subcommand",
			args: []string{"fleet"},
			want: "fleet requires a subcommand",
		},
		{
			name: "fleet bad subcommand",
			args: []string{"fleet", "bogus"},
			want: `unsupported fleet subcommand "bogus"`,
		},
		{
			name: "indexes missing branch",
			args: []string{"indexes"},
			want: "indexes requires --branch",
		},
		{
			name: "fleet health missing repo list",
			args: []string{"fleet", "health", "--branch", "main"},
			want: "fleet health requires --repo-list",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := run(tc.args)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q in %q", tc.want, err.Error())
			}
		})
	}
}

func TestNormalizeFormat(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		jsonFlag bool
		want     string
		wantErr  string
	}{
		{name: "lowercases text", format: "TEXT", want: "text"},
		{name: "trims markdown", format: " markdown ", want: "markdown"},
		{name: "json flag wins", format: "tsv", jsonFlag: true, want: "json"},
		{name: "bad format", format: "yaml", wantErr: `unsupported format "yaml"`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalizeFormat(tc.format, tc.jsonFlag)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected %q in %q", tc.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestRunAgeEndToEnd(t *testing.T) {
	repo := sampleCLIRepo(t)

	out := captureStdout(t, func() {
		if err := run([]string{"age", "--repo", repo.Dir, "--branch", "jj", "--format", "text"}); err != nil {
			t.Fatalf("run age: %v", err)
		}
	})

	for _, want := range []string{"TODO-ravud", "TODO/TODO.md", "Legacy task"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in %q", want, out)
		}
	}
}

func TestRunDriftEndToEnd(t *testing.T) {
	repo := sampleCLIRepo(t)

	out := captureStdout(t, func() {
		if err := run([]string{"drift", "--repo", repo.Dir, "--branch-a", "main", "--branch-b", "jj", "--format", "tsv"}); err != nil {
			t.Fatalf("run drift: %v", err)
		}
	})

	for _, want := range []string{"kind\tvalue\tdetails", "only_in_jj\tTODO-ravud", "completed_only_in_jj\tTODO-binap"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in %q", want, out)
		}
	}
}

func TestRunLintEndToEnd(t *testing.T) {
	repo := sampleCLIRepo(t)

	out := captureStdout(t, func() {
		if err := run([]string{"lint", "--repo", repo.Dir, "--branch", "jj", "--format", "markdown"}); err != nil {
			t.Fatalf("run lint: %v", err)
		}
	})

	for _, want := range []string{"## Lint Report", "checked_parent_open_subtask", "orphan_detail_file"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in %q", want, out)
		}
	}
}

func TestRunHealthEndToEnd(t *testing.T) {
	repo := sampleCLIRepo(t)

	out := captureStdout(t, func() {
		if err := run([]string{"health", "--repo", repo.Dir, "--branch", "jj", "--compare", "main", "--json"}); err != nil {
			t.Fatalf("run health: %v", err)
		}
	})

	for _, want := range []string{`"repo":`, `"branch": "jj"`, `"open_todos": 2`, `"completed_todos": 1`, `"total_difference_rows": 3`, `"status": "warning"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in %q", want, out)
		}
	}
}

func TestRunIndexesEndToEnd(t *testing.T) {
	repo := sampleMultiIndexRepo(t)

	out := captureStdout(t, func() {
		if err := run([]string{"indexes", "--repo", repo.Dir, "--branch", "main", "--format", "text"}); err != nil {
			t.Fatalf("run indexes: %v", err)
		}
	})

	for _, want := range []string{"TODO/TODO.md", "protocols/wire-lab.d/TODO/TODO.md"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in %q", want, out)
		}
	}
}

func TestRunHealthAllIndexesEndToEnd(t *testing.T) {
	repo := sampleMultiIndexRepo(t)

	out := captureStdout(t, func() {
		if err := run([]string{"health", "--repo", repo.Dir, "--branch", "main", "--all-indexes", "--format", "text"}); err != nil {
			t.Fatalf("run health all indexes: %v", err)
		}
	})

	for _, want := range []string{"Discovered indexes: 2", "Index summaries:", "protocols/wire-lab.d/TODO/TODO.md"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in %q", want, out)
		}
	}
}

func TestRunHealthAllIndexesCompareEndToEnd(t *testing.T) {
	repo := sampleMultiIndexCompareRepo(t)

	out := captureStdout(t, func() {
		if err := run([]string{"health", "--repo", repo.Dir, "--branch", "main", "--all-indexes", "--compare", "jj", "--format", "text"}); err != nil {
			t.Fatalf("run health all indexes compare: %v", err)
		}
	})

	for _, want := range []string{"Compare branch: jj", "Repo-wide drift rows:", "Indexes only in jj:", "drift=3"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in %q", want, out)
		}
	}
}

func TestRunLintWithNestedIndexEndToEnd(t *testing.T) {
	repo := sampleNestedIndexRepo(t)

	out := captureStdout(t, func() {
		if err := run([]string{"lint", "--repo", repo.Dir, "--branch", "main", "--index", "protocols/wire-lab.d/TODO/TODO.md", "--format", "text"}); err != nil {
			t.Fatalf("run lint: %v", err)
		}
	})

	if !strings.Contains(out, "WARNING orphan_detail_file") {
		t.Fatalf("expected orphan detail warning in %q", out)
	}
}

func TestRunFleetHealthEndToEnd(t *testing.T) {
	repoA := sampleCLIRepo(t)
	repoB := sampleMultiIndexCompareRepo(t)
	listDir := t.TempDir()
	repoList := filepath.Join(listDir, "repos.txt")
	content := strings.Join([]string{
		repoA.Dir,
		repoB.Dir,
		filepath.Join(listDir, "missing-repo"),
	}, "\n")
	if err := os.WriteFile(repoList, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	out := captureStdout(t, func() {
		if err := run([]string{"fleet", "health", "--repo-list", repoList, "--branch", "main", "--all-indexes", "--compare", "jj", "--format", "text"}); err != nil {
			t.Fatalf("run fleet health: %v", err)
		}
	})

	for _, want := range []string{"Repo list:", "Repos: 3", "Successful repos: 2", "Repo errors: 1", "mode=all-indexes", "missing-repo"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in %q", want, out)
		}
	}
}

func sampleCLIRepo(t *testing.T) *testrepo.Repo {
	t.Helper()

	repo := testrepo.New(t)
	repo.Write("TODO/TODO.md", "# TODO Index\n\n- [ ] TODO-binap - Lock outline (`TODO/TODO-binap.md`)\n001 - Legacy task\n")
	repo.Write("TODO/TODO-binap.md", "# TODO-binap\n\n- [ ] binap.1 First subtask\n")
	repo.Commit("Seed main", "2026-01-01T00:00:00Z")

	repo.CheckoutNew("jj")
	repo.Write("TODO/TODO.md", "# TODO Index\n\n- [x] TODO-binap - Lock outline (`TODO/TODO-binap.md`)\n- [ ] TODO-ravud - New work (`TODO/TODO-ravud.md`)\n001 - Legacy task\n")
	repo.Write("TODO/TODO-ravud.md", "# TODO-ravud\n\n- [ ] ravud.1 Branch-only subtask\n")
	repo.Write("TODO/TODO-orphan.md", "# TODO-orphan\n\n- [ ] orphan.1 Unreferenced detail file\n")
	repo.Commit("Update jj", "2026-01-02T00:00:00Z")

	return repo
}

func sampleNestedIndexRepo(t *testing.T) *testrepo.Repo {
	t.Helper()

	repo := testrepo.New(t)
	repo.Write("protocols/wire-lab.d/TODO/TODO.md", "# TODO queue\n\n| Handle | Mint date | Title | Prior alias |\n|---|---|---|---|\n| [TODO-hipak](./TODO-hipak-local.md) | 2026-05-25 | Local title | — |\n| [TODO-bisur](../../../simulations/SIM-rakot-group-session/protocols/group-session.d/TODO/TODO-bisur-group-transport-envelope.md) | 2026-05-01 | Cross-tree title | — |\n")
	repo.Write("protocols/wire-lab.d/TODO/TODO-hipak-local.md", "# TODO-hipak\n\n## Status\n\nRunning.\n\n## Subtasks\n\n- [ ] hipak.1 First subtask\n")
	repo.Write("protocols/wire-lab.d/TODO/TODO-orphan-local.md", "# TODO-orphan\n\n## Status\n\nPlanned.\n")
	repo.Write("simulations/SIM-rakot-group-session/protocols/group-session.d/TODO/TODO-bisur-group-transport-envelope.md", "# TODO-bisur\n\n## Status\n\nImplemented.\n\n## Subtasks\n\n- [x] bisur.1 Done subtask\n")
	repo.Commit("Seed nested index", "2026-01-01T00:00:00Z")

	return repo
}

func sampleMultiIndexRepo(t *testing.T) *testrepo.Repo {
	t.Helper()

	repo := testrepo.New(t)
	repo.Write("TODO/TODO.md", "# TODO Index\n\n- [ ] TODO-roota - Root task (`TODO/TODO-roota.md`)\n")
	repo.Write("TODO/TODO-roota.md", "# TODO-roota\n\n- [ ] roota.1 Root subtask\n")
	repo.Write("protocols/wire-lab.d/TODO/TODO.md", "# TODO queue\n\n| Handle | Mint date | Title | Prior alias |\n|---|---|---|---|\n| [TODO-hipak](./TODO-hipak-local.md) | 2026-05-25 | Local title | — |\n")
	repo.Write("protocols/wire-lab.d/TODO/TODO-hipak-local.md", "# TODO-hipak\n\n## Status\n\nRunning.\n\n## Subtasks\n\n- [ ] hipak.1 First subtask\n")
	repo.Commit("Seed multi index", "2026-01-01T00:00:00Z")

	return repo
}

func sampleMultiIndexCompareRepo(t *testing.T) *testrepo.Repo {
	t.Helper()

	repo := testrepo.New(t)
	repo.Write("TODO/TODO.md", "# TODO Index\n\n- [ ] TODO-roota - Root task (`TODO/TODO-roota.md`)\n")
	repo.Write("TODO/TODO-roota.md", "# TODO-roota\n\n- [ ] roota.1 Root subtask\n")
	repo.Write("protocols/wire-lab.d/TODO/TODO.md", "# TODO queue\n\n| Handle | Mint date | Title | Prior alias |\n|---|---|---|---|\n| [TODO-hipak](./TODO-hipak-local.md) | 2026-05-25 | Local title | — |\n")
	repo.Write("protocols/wire-lab.d/TODO/TODO-hipak-local.md", "# TODO-hipak\n\n## Status\n\nRunning.\n\n## Subtasks\n\n- [ ] hipak.1 First subtask\n")
	repo.Write("simulations/SIM-alpha/TODO/TODO.md", "# TODO Index\n\n- [ ] TODO-alpha - Alpha task (`simulations/SIM-alpha/TODO/TODO-alpha.md`)\n")
	repo.Write("simulations/SIM-alpha/TODO/TODO-alpha.md", "# TODO-alpha\n\n- [ ] alpha.1 Alpha subtask\n")
	repo.Commit("Seed multi index compare", "2026-01-01T00:00:00Z")

	repo.CheckoutNew("jj")
	repo.Write("TODO/TODO.md", "# TODO Index\n\n- [x] TODO-roota - Root task (`TODO/TODO-roota.md`)\n")
	repo.Write("protocols/wire-lab.d/TODO/TODO.md", "# TODO queue\n\n| Handle | Mint date | Title | Prior alias |\n|---|---|---|---|\n| [TODO-hipak](./TODO-hipak-local.md) | 2026-05-25 | Local title updated | — |\n")
	repo.Write("protocols/wire-lab.d/TODO/TODO-hipak-local.md", "# TODO-hipak\n\n## Status\n\nImplemented.\n\n## Subtasks\n\n- [x] hipak.1 First subtask\n")
	repo.Write("simulations/SIM-beta/TODO/TODO.md", "# TODO Index\n\n- [ ] TODO-beta - Beta task (`simulations/SIM-beta/TODO/TODO-beta.md`)\n")
	repo.Write("simulations/SIM-beta/TODO/TODO-beta.md", "# TODO-beta\n\n- [ ] beta.1 Beta subtask\n")
	repo.Commit("Diverge multi indexes", "2026-01-02T00:00:00Z")

	return repo
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close reader: %v", err)
	}
	return buf.String()
}
