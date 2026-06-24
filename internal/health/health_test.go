package health

import (
	"testing"
	"time"

	"todo-report/internal/model"
)

func TestBuildSummarizesHealthReport(t *testing.T) {
	snapshot := model.Snapshot{
		RepoName:     "coordination",
		Branch:       "jj",
		IndexFile:    "TODO/TODO.md",
		TodoRoot:     "TODO",
		Items:        []model.TodoItem{{TodoID: "TODO-old", Status: model.StatusOpen}, {TodoID: "TODO-done", Status: model.StatusCompleted}},
		OrphanDetail: []string{"TODO/TODO-orphan.md"},
	}
	ages := []model.AgeRecord{
		{Todo: model.TodoItem{TodoID: "TODO-old", Status: model.StatusOpen, Title: "Old open"}, FirstSeen: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), AgeDays: 60},
		{Todo: model.TodoItem{TodoID: "TODO-done", Status: model.StatusCompleted, Title: "Done"}, FirstSeen: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), AgeDays: 30},
	}
	findings := []model.LintFinding{
		{Severity: "error", Code: "missing_detail_file"},
		{Severity: "warning", Code: "orphan_detail_file"},
		{Severity: "warning", Code: "orphan_detail_file"},
	}

	report := Build(snapshot, ages, findings, nil, "")
	if report.Status != "error" {
		t.Fatalf("expected error status, got %q", report.Status)
	}
	if report.OpenTODOs != 1 || report.CompletedTODOs != 1 {
		t.Fatalf("unexpected todo counts %#v", report)
	}
	if report.OldestOpen == nil || report.OldestOpen.Todo.TodoID != "TODO-old" {
		t.Fatalf("unexpected oldest open %#v", report.OldestOpen)
	}
	if len(report.OldestOpenItems) != 1 {
		t.Fatalf("unexpected oldest open items %#v", report.OldestOpenItems)
	}
	if len(report.FindingSummary) != 2 {
		t.Fatalf("expected grouped finding summary, got %#v", report.FindingSummary)
	}
	if report.FindingSummary[0].Code != "missing_detail_file" || report.FindingSummary[1].Count != 2 {
		t.Fatalf("unexpected finding summary %#v", report.FindingSummary)
	}
}

func TestBuildMultiAggregatesReports(t *testing.T) {
	reports := []model.HealthReport{
		{IndexFile: "TODO/TODO.md", Status: "warning", OpenTODOs: 2, CompletedTODOs: 1, LintWarnings: 3, Drift: &model.DriftResult{TotalDifferenceRows: 2}},
		{IndexFile: "protocols/wire-lab.d/TODO/TODO.md", Status: "error", OpenTODOs: 5, CompletedTODOs: 2, LintErrors: 1},
	}

	report := BuildMulti("wire-lab", "main", "jj", reports, []string{"TODO/TODO.md"}, []string{"simulations/SIM-beta/TODO/TODO.md"})
	if report.Status != "error" {
		t.Fatalf("expected error status, got %q", report.Status)
	}
	if report.OpenTODOs != 7 || report.CompletedTODOs != 3 {
		t.Fatalf("unexpected totals %#v", report)
	}
	if report.IndexesWithErrors != 1 || report.IndexesWithWarning != 1 {
		t.Fatalf("unexpected index severity counts %#v", report)
	}
	if report.DriftItems != 2 || report.IndexesWithDrift != 1 {
		t.Fatalf("unexpected drift aggregation %#v", report)
	}
	if len(report.IndexesOnlyInBranch) != 1 || len(report.IndexesOnlyInCompare) != 1 {
		t.Fatalf("unexpected branch-only index lists %#v", report)
	}
	if len(report.IndexFiles) != 3 || report.IndexFiles[0] != "TODO/TODO.md" {
		t.Fatalf("unexpected index files %#v", report.IndexFiles)
	}
}
