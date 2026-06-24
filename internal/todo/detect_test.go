package todo

import (
	"testing"

	"todo-report/internal/gitrepo"
	"todo-report/internal/testrepo"
)

func TestDetectSnapshotCompatible(t *testing.T) {
	repoDir := testrepo.New(t)
	repoDir.Write("TODO/TODO.md", "# TODO Index\n\n- [ ] TODO-binap - Lock outline (`TODO/TODO-binap.md`)\n001 - Legacy task\njirin - Bare proquint task (`TODO/TODO-jirin.md`)\n014-change-review-gate.md Change review gate\n")
	repoDir.Write("TODO/TODO-binap.md", "# TODO-binap\n\n- [ ] binap.1 First subtask\n- [~] 1. Approximate\n")
	repoDir.Write("TODO/TODO-jirin.md", "# TODO-jirin\n\n- [x] jirin.1 Done\n")
	repoDir.Write("TODO/014-change-review-gate.md", "# TODO-014\n\n- [ ] 014.1 First\n")
	repoDir.Commit("Seed detect repo", "2026-01-01T00:00:00Z")

	repo, err := gitrepo.Open(repoDir.Dir)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := LoadSnapshot(repo, "main", "TODO/TODO.md")
	if err != nil {
		t.Fatal(err)
	}

	report := DetectSnapshot(snapshot)
	if report.Compatibility != "compatible_with_warnings" {
		t.Fatalf("expected compatible_with_warnings, got %#v", report)
	}
	if !contains(report.TopLevelIDStyles, "proquint") || !contains(report.TopLevelIDStyles, "numeric_legacy") || !contains(report.TopLevelIDStyles, "bare_proquint") || !contains(report.TopLevelIDStyles, "filename_stem") {
		t.Fatalf("unexpected top-level styles %#v", report.TopLevelIDStyles)
	}
	if !contains(report.Features, "approximate_checkboxes") || !contains(report.Features, "trailing_dot_subtask_ids") {
		t.Fatalf("unexpected features %#v", report.Features)
	}
}

func TestDetectSnapshotUnsupported(t *testing.T) {
	repoDir := testrepo.New(t)
	repoDir.Write("TODO/TODO.md", "# TODO Index\n\n- [q] TODO-binap - Lock outline (TODO/TODO-binap.md)\n")
	repoDir.Write("TODO/TODO-binap.md", "# TODO-binap\n\n- [ ] (bad.id) Bad\n")
	repoDir.Commit("Seed incompatible detect repo", "2026-01-01T00:00:00Z")

	repo, err := gitrepo.Open(repoDir.Dir)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := LoadSnapshot(repo, "main", "TODO/TODO.md")
	if err != nil {
		t.Fatal(err)
	}

	report := DetectSnapshot(snapshot)
	if report.Compatibility != "unsupported" {
		t.Fatalf("expected unsupported, got %#v", report)
	}
	if len(report.StyleFindings) == 0 {
		t.Fatalf("expected style findings, got %#v", report)
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
