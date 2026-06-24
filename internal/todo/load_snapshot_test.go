package todo

import (
	"testing"

	"todo-report/internal/gitrepo"
	"todo-report/internal/model"
	"todo-report/internal/testrepo"
)

func TestLoadSnapshotWithNestedIndexAndRelativeLinks(t *testing.T) {
	repoDir := testrepo.New(t)
	repoDir.Write("protocols/wire-lab.d/TODO/TODO.md", "# TODO queue\n\n| Handle | Mint date | Title | Prior alias |\n|---|---|---|---|\n| [TODO-hipak](./TODO-hipak-local.md) | 2026-05-25 | Local title | — |\n| [TODO-bisur](../../../simulations/SIM-rakot-group-session/protocols/group-session.d/TODO/TODO-bisur-group-transport-envelope.md) | 2026-05-01 | Cross-tree title | — |\n")
	repoDir.Write("protocols/wire-lab.d/TODO/TODO-hipak-local.md", "# TODO-hipak\n\n## Status\n\nRunning.\n\n## Subtasks\n\n- [ ] hipak.1 First subtask\n")
	repoDir.Write("protocols/wire-lab.d/TODO/TODO-orphan-local.md", "# TODO-orphan\n\n## Status\n\nPlanned.\n")
	repoDir.Write("simulations/SIM-rakot-group-session/protocols/group-session.d/TODO/TODO-bisur-group-transport-envelope.md", "# TODO-bisur\n\n## Status\n\nImplemented.\n\n## Subtasks\n\n- [x] bisur.1 Done subtask\n")
	repoDir.Commit("Seed nested index", "2026-01-01T00:00:00Z")

	repo, err := gitrepo.Open(repoDir.Dir)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := LoadSnapshot(repo, "main", "protocols/wire-lab.d/TODO/TODO.md")
	if err != nil {
		t.Fatal(err)
	}

	if snapshot.IndexFile != "protocols/wire-lab.d/TODO/TODO.md" {
		t.Fatalf("unexpected index file %q", snapshot.IndexFile)
	}
	if snapshot.TodoRoot != "protocols/wire-lab.d/TODO" {
		t.Fatalf("unexpected todo root %q", snapshot.TodoRoot)
	}
	if len(snapshot.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(snapshot.Items))
	}
	if snapshot.ItemByID["TODO-hipak"].Status != model.StatusOpen {
		t.Fatalf("expected TODO-hipak to be open, got %q", snapshot.ItemByID["TODO-hipak"].Status)
	}
	if snapshot.ItemByID["TODO-bisur"].Status != model.StatusCompleted {
		t.Fatalf("expected TODO-bisur to be completed, got %q", snapshot.ItemByID["TODO-bisur"].Status)
	}
	if len(snapshot.SubtasksByParent["TODO-hipak"]) != 1 {
		t.Fatalf("expected TODO-hipak subtasks, got %#v", snapshot.SubtasksByParent["TODO-hipak"])
	}
	if len(snapshot.SubtasksByParent["TODO-bisur"]) != 1 {
		t.Fatalf("expected TODO-bisur subtasks, got %#v", snapshot.SubtasksByParent["TODO-bisur"])
	}
	if len(snapshot.OrphanDetail) != 1 || snapshot.OrphanDetail[0] != "protocols/wire-lab.d/TODO/TODO-orphan-local.md" {
		t.Fatalf("unexpected orphan detail files %#v", snapshot.OrphanDetail)
	}
}
