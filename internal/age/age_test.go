package age

import (
	"testing"

	"todo-report/internal/gitrepo"
	"todo-report/internal/testrepo"
	"todo-report/internal/todo"
)

func TestComputeUsesFirstAppearanceInBranchHistory(t *testing.T) {
	repoDir := testrepo.New(t)
	repoDir.Write("TODO/TODO.md", "# TODO Index\n\n001 - First task\n")
	repoDir.Commit("Add first todo", "2026-01-01T00:00:00Z")

	repoDir.Write("TODO/TODO.md", "# TODO Index\n\n001 - First task\n- [ ] TODO-binap - Lock outline (`TODO/TODO-binap.md`)\n")
	repoDir.Write("TODO/TODO-binap.md", "# TODO-binap\n\n- [ ] binap.1 First subtask\n")
	repoDir.Commit("Add second todo", "2026-02-01T00:00:00Z")

	repo, err := gitrepo.Open(repoDir.Dir)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := todo.LoadSnapshot(repo, "main")
	if err != nil {
		t.Fatal(err)
	}

	records, err := Compute(repo, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	got := map[string]string{}
	for _, record := range records {
		got[record.Todo.TodoID] = record.FirstSeen.Format("2006-01-02")
	}
	if got["001"] != "2026-01-01" {
		t.Fatalf("expected 001 first seen on 2026-01-01, got %s", got["001"])
	}
	if got["TODO-binap"] != "2026-02-01" {
		t.Fatalf("expected TODO-binap first seen on 2026-02-01, got %s", got["TODO-binap"])
	}
}
