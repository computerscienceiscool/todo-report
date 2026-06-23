package gitrepo

import (
	"path/filepath"
	"strings"
	"testing"

	"todo-report/internal/testrepo"
)

func TestOpenFromSubdirectory(t *testing.T) {
	repoDir := testrepo.New(t)
	repoDir.Write("TODO/TODO.md", "# TODO Index\n\n001 - First task\n")
	repoDir.Commit("Seed repo", "2026-01-01T00:00:00Z")

	subdir := filepath.Join(repoDir.Dir, "nested", "child")
	repoDir.Write("nested/child/.keep", "")

	repo, err := Open(subdir)
	if err != nil {
		t.Fatalf("open repo: %v", err)
	}
	if repo.Root != repoDir.Dir {
		t.Fatalf("expected root %q, got %q", repoDir.Dir, repo.Root)
	}
	if repo.Name == "" {
		t.Fatal("expected repo name")
	}
}

func TestRepoOperations(t *testing.T) {
	repoDir := testrepo.New(t)
	repoDir.Write("TODO/TODO.md", "# TODO Index\n\n001 - First task\n")
	repoDir.Commit("First commit", "2026-01-01T00:00:00Z")

	repoDir.Write("TODO/TODO.md", "# TODO Index\n\n001 - First task\n- [ ] TODO-binap - Lock outline (`TODO/TODO-binap.md`)\n")
	repoDir.Write("TODO/TODO-binap.md", "# TODO-binap\n\n- [ ] binap.1 First subtask\n")
	repoDir.Commit("Second commit", "2026-02-01T00:00:00Z")

	repo, err := Open(repoDir.Dir)
	if err != nil {
		t.Fatalf("open repo: %v", err)
	}

	commit, err := repo.BranchCommit("main")
	if err != nil {
		t.Fatalf("branch commit: %v", err)
	}
	if len(commit) != 40 {
		t.Fatalf("expected 40-char commit hash, got %q", commit)
	}

	content, err := repo.ShowFile("main", "TODO/TODO.md")
	if err != nil {
		t.Fatalf("show file: %v", err)
	}
	if want := "TODO-binap"; !strings.Contains(content, want) {
		t.Fatalf("expected %q in %q", want, content)
	}

	files, err := repo.ListFiles("main", "TODO")
	if err != nil {
		t.Fatalf("list files: %v", err)
	}
	if len(files) != 2 || files[0] != "TODO/TODO-binap.md" || files[1] != "TODO/TODO.md" {
		t.Fatalf("unexpected files: %#v", files)
	}

	history, err := repo.ReverseLog("main", "TODO/TODO.md")
	if err != nil {
		t.Fatalf("reverse log: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(history))
	}
	if !history[0].When.Before(history[1].When) {
		t.Fatalf("expected reverse chronological order, got %#v", history)
	}
}
