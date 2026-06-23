package drift

import (
	"testing"

	"todo-report/internal/gitrepo"
	"todo-report/internal/testrepo"
	"todo-report/internal/todo"
)

func TestCompareTracksTopLevelAndSubtaskDrift(t *testing.T) {
	repoDir := testrepo.New(t)
	repoDir.Write("TODO/TODO.md", "# TODO Index\n\n- [ ] TODO-binap - Lock outline (`TODO/TODO-binap.md`)\n001 - Old style task\n")
	repoDir.Write("TODO/TODO-binap.md", "# TODO-binap\n\n- [ ] binap.1 First\n")
	repoDir.Commit("Seed main", "2026-01-01T00:00:00Z")

	repoDir.CheckoutNew("jj")
	repoDir.Write("TODO/TODO.md", "# TODO Index\n\n- [x] TODO-binap - Lock outline (`TODO/TODO-binap.md`)\n001 - Old style task\n- [ ] TODO-ravud - New branch todo (`TODO/TODO-ravud.md`)\n")
	repoDir.Write("TODO/TODO-binap.md", "# TODO-binap\n\n- [x] binap.1 First\n- [ ] binap.2 Added later\n")
	repoDir.Write("TODO/TODO-ravud.md", "# TODO-ravud\n\n- [ ] ravud.1 Branch only\n")
	repoDir.Commit("Update branch", "2026-01-02T00:00:00Z")

	repo, err := gitrepo.Open(repoDir.Dir)
	if err != nil {
		t.Fatal(err)
	}
	mainSnapshot, err := todo.LoadSnapshot(repo, "main")
	if err != nil {
		t.Fatal(err)
	}
	jjSnapshot, err := todo.LoadSnapshot(repo, "jj")
	if err != nil {
		t.Fatal(err)
	}

	result := Compare(mainSnapshot, jjSnapshot)
	if len(result.OnlyInB) != 1 || result.OnlyInB[0] != "TODO-ravud" {
		t.Fatalf("expected TODO-ravud only in jj, got %#v", result.OnlyInB)
	}
	if len(result.CompletedOnlyInB) != 1 || result.CompletedOnlyInB[0] != "TODO-binap" {
		t.Fatalf("expected TODO-binap completed only in jj, got %#v", result.CompletedOnlyInB)
	}
	if len(result.SubtaskOnlyInB) != 2 {
		t.Fatalf("expected 2 subtasks only in jj, got %#v", result.SubtaskOnlyInB)
	}
	if len(result.SubtaskCompletedB) != 1 || result.SubtaskCompletedB[0] != "TODO-binap/binap.1" {
		t.Fatalf("expected binap.1 completed only in jj, got %#v", result.SubtaskCompletedB)
	}
}
