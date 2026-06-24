package lint

import (
	"testing"

	"todo-report/internal/gitrepo"
	"todo-report/internal/testrepo"
	"todo-report/internal/todo"
)

func TestRunFindsOrphanAndIndexDoneDetailOpenWarnings(t *testing.T) {
	repoDir := testrepo.New(t)
	repoDir.Write("TODO/TODO.md", "# TODO Index\n\n- [x] TODO-binap - Lock outline (`TODO/TODO-binap.md`)\n")
	repoDir.Write("TODO/TODO-binap.md", "# TODO-binap\n\n- [ ] binap.1 Still open\n")
	repoDir.Write("TODO/TODO-orphan.md", "# TODO-orphan\n\n- [ ] orphan.1 Unreferenced\n")
	repoDir.Commit("Seed repo", "2026-01-01T00:00:00Z")

	repo, err := gitrepo.Open(repoDir.Dir)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := todo.LoadSnapshot(repo, "main", "TODO/TODO.md")
	if err != nil {
		t.Fatal(err)
	}
	findings := Run(snapshot)

	codes := map[string]bool{}
	for _, finding := range findings {
		codes[finding.Code] = true
	}
	if !codes["index_done_detail_open"] {
		t.Fatalf("expected index_done_detail_open in %#v", findings)
	}
	if !codes["orphan_detail_file"] {
		t.Fatalf("expected orphan_detail_file in %#v", findings)
	}
}

func TestRunFindsIndexOpenDetailCompleteWarning(t *testing.T) {
	repoDir := testrepo.New(t)
	repoDir.Write("TODO/TODO.md", "# TODO Index\n\n- [ ] jirin - Bootstrap work (`TODO/TODO-jirin.md`)\n")
	repoDir.Write("TODO/TODO-jirin.md", "# TODO-jirin\n\n## Status\n\nImplemented.\n\n- [x] jirin.1 Done\n- [x] jirin.2 Also done\n")
	repoDir.Commit("Seed repo", "2026-01-01T00:00:00Z")

	repo, err := gitrepo.Open(repoDir.Dir)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := todo.LoadSnapshot(repo, "main", "TODO/TODO.md")
	if err != nil {
		t.Fatal(err)
	}
	findings := Run(snapshot)

	codes := map[string]bool{}
	for _, finding := range findings {
		codes[finding.Code] = true
	}
	if !codes["index_open_detail_complete"] {
		t.Fatalf("expected index_open_detail_complete in %#v", findings)
	}
}
