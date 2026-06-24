package fleet

import (
	"os"
	"path/filepath"
	"testing"

	"todo-report/internal/model"
)

func TestLoadRepoList(t *testing.T) {
	dir := t.TempDir()
	listPath := filepath.Join(dir, "repos.txt")
	relRepo := filepath.Join("nested", "repo")
	if err := os.WriteFile(listPath, []byte("# repos\n\n"+relRepo+"\n./second\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	repos, err := LoadRepoList(listPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %#v", repos)
	}
	if repos[0] != filepath.Join(dir, relRepo) {
		t.Fatalf("unexpected first repo %q", repos[0])
	}
}

func TestBuildHealthReport(t *testing.T) {
	entries := []model.FleetHealthEntry{
		{Repo: "a", RepoPath: "/tmp/a", Status: "warning", IndexMode: "all-indexes", IndexCount: 2, OpenTODOs: 3, CompletedTODOs: 1, LintWarnings: 2, DriftItems: 4},
		{Repo: "b", RepoPath: "/tmp/b", Status: "error", Error: "open repo failed", IndexMode: "single-index"},
	}

	report := BuildHealthReport("main", "jj", "/tmp/repos.txt", entries)
	if report.Status != "error" {
		t.Fatalf("expected error status, got %q", report.Status)
	}
	if report.RepoCount != 2 || report.SuccessCount != 1 || report.ErrorCount != 1 {
		t.Fatalf("unexpected fleet counts %#v", report)
	}
	if report.OpenTODOs != 3 || report.LintWarnings != 2 || report.DriftItems != 4 {
		t.Fatalf("unexpected fleet totals %#v", report)
	}
}
