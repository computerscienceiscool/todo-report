package todo

import (
	"testing"

	"todo-report/internal/gitrepo"
	"todo-report/internal/testrepo"
)

func TestDiscoverIndexesFindsRootAndNestedIndexes(t *testing.T) {
	repoDir := testrepo.New(t)
	repoDir.Write("TODO/TODO.md", "# TODO Index\n\n001 - Root task\n")
	repoDir.Write("protocols/wire-lab.d/TODO/TODO.md", "# Nested TODO Index\n\n001 - Nested task\n")
	repoDir.Write("protocols/wire-lab.d/TODO/TODO-bisur.md", "# TODO-bisur\n")
	repoDir.Write("notes/TODO.md", "# Not a TODO root\n")
	repoDir.Commit("Seed indexes", "2026-01-01T00:00:00Z")

	repo, err := gitrepo.Open(repoDir.Dir)
	if err != nil {
		t.Fatal(err)
	}

	indexes, err := DiscoverIndexes(repo, "main")
	if err != nil {
		t.Fatal(err)
	}

	if len(indexes) != 2 {
		t.Fatalf("expected 2 indexes, got %#v", indexes)
	}
	if indexes[0] != "TODO/TODO.md" || indexes[1] != "protocols/wire-lab.d/TODO/TODO.md" {
		t.Fatalf("unexpected indexes %#v", indexes)
	}
}
