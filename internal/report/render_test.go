package report

import (
	"strings"
	"testing"
	"time"

	"todo-report/internal/model"
)

func TestRenderAgeFormats(t *testing.T) {
	records := []model.AgeRecord{
		{
			Todo: model.TodoItem{
				TodoID:     "TODO-binap",
				Title:      "Lock outline",
				Status:     model.StatusOpen,
				SourceFile: "TODO/TODO.md",
				DetailFile: "TODO/TODO-binap.md",
			},
			FirstSeen: time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC),
			AgeDays:   45,
		},
	}

	tests := []struct {
		format string
		want   []string
	}{
		{format: "text", want: []string{"TODO-binap", "45 days", "Lock outline"}},
		{format: "markdown", want: []string{"## Age Report", "- [ ] `TODO-binap`"}},
		{format: "json", want: []string{`"todo_id": "TODO-binap"`, `"age_days": 45`}},
		{format: "tsv", want: []string{"todo_id\tstatus\tage_days", "TODO-binap\topen\t45"}},
	}

	for _, tc := range tests {
		t.Run(tc.format, func(t *testing.T) {
			out, err := RenderAge(records, tc.format)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, want := range tc.want {
				if !strings.Contains(out, want) {
					t.Fatalf("expected %q in output %q", want, out)
				}
			}
		})
	}
}

func TestRenderLintFormats(t *testing.T) {
	snapshot := sampleSnapshot()
	findings := []model.LintFinding{
		{
			Severity: "error",
			Code:     "missing_detail_file",
			TodoID:   "TODO-binap",
			File:     "TODO/TODO.md",
			Line:     12,
			Message:  "Detail file does not exist.",
		},
	}

	tests := []struct {
		format string
		want   []string
	}{
		{format: "text", want: []string{"ERROR missing_detail_file", "TODO/TODO.md:12"}},
		{format: "markdown", want: []string{"## Lint Report", "`missing_detail_file`", "`TODO-binap`"}},
		{format: "json", want: []string{`"code": "missing_detail_file"`}},
		{format: "tsv", want: []string{"severity\tcode\ttodo_id", "error\tmissing_detail_file\tTODO-binap"}},
	}

	for _, tc := range tests {
		t.Run(tc.format, func(t *testing.T) {
			out, err := RenderLint(snapshot, findings, tc.format)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, want := range tc.want {
				if !strings.Contains(out, want) {
					t.Fatalf("expected %q in output %q", want, out)
				}
			}
		})
	}
}

func TestRenderDriftFormats(t *testing.T) {
	result := model.DriftResult{
		RepoName:          "coordination",
		BranchA:           "main",
		BranchB:           "jj",
		OnlyInA:           []string{"001"},
		OnlyInB:           []string{"TODO-binap"},
		SubtaskCompletedB: []string{"TODO-binap/binap.1"},
		OtherDifferences: []model.DriftChange{
			{Kind: "title_changed", TodoID: "TODO-binap", Details: "A <> B"},
		},
	}

	tests := []struct {
		format string
		want   []string
	}{
		{format: "text", want: []string{"Only in main:", "TODO-binap/binap.1"}},
		{format: "markdown", want: []string{"## Drift Report", "### Only in main", "`TODO-binap/binap.1`"}},
		{format: "json", want: []string{`"branch_a": "main"`, `"TODO-binap"`}},
		{format: "tsv", want: []string{"kind\tvalue\tdetails", "only_in_main\t001\t", "title_changed\tTODO-binap\tA <> B"}},
	}

	for _, tc := range tests {
		t.Run(tc.format, func(t *testing.T) {
			out, err := RenderDrift(result, tc.format)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, want := range tc.want {
				if !strings.Contains(out, want) {
					t.Fatalf("expected %q in output %q", want, out)
				}
			}
		})
	}
}

func TestRenderHealthFormats(t *testing.T) {
	snapshot := sampleSnapshot()
	ages := []model.AgeRecord{
		{
			Todo: model.TodoItem{
				TodoID: "TODO-binap",
				Status: model.StatusOpen,
				Title:  "Lock outline",
			},
			AgeDays: 45,
		},
	}
	findings := []model.LintFinding{
		{Severity: "error", Code: "missing_detail_file", File: "TODO/TODO.md", Line: 12},
		{Severity: "warning", Code: "orphan_detail_file", File: "TODO/TODO-orphan.md", Line: 1},
	}
	driftResult := &model.DriftResult{BranchA: "main", BranchB: "jj", TotalDifferenceRows: 4}

	tests := []struct {
		format string
		want   []string
	}{
		{format: "text", want: []string{"Repo: coordination", "Branch drift items: 4"}},
		{format: "markdown", want: []string{"## Health Report", "## Drift Report", "## Lint Report"}},
		{format: "json", want: []string{`"open_todos": 1`, `"lint_errors": 1`, `"total_difference_rows": 4`}},
		{format: "tsv", want: []string{"key\tvalue", "repo\tcoordination", "drift_items\t4"}},
	}

	for _, tc := range tests {
		t.Run(tc.format, func(t *testing.T) {
			out, err := RenderHealth(snapshot, ages, findings, driftResult, tc.format)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, want := range tc.want {
				if !strings.Contains(out, want) {
					t.Fatalf("expected %q in output %q", want, out)
				}
			}
		})
	}
}

func TestRenderUnsupportedFormat(t *testing.T) {
	_, err := RenderAge(nil, "yaml")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), `unsupported format "yaml"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHelperFunctions(t *testing.T) {
	keys := SortedKeys(map[string]int{"b": 2, "a": 1})
	if len(keys) != 2 || keys[0] != "a" || keys[1] != "b" {
		t.Fatalf("unexpected keys: %#v", keys)
	}

	got := CompactLines([]string{" one ", "", " two"})
	if got != "one\ntwo\n" {
		t.Fatalf("unexpected compact output %q", got)
	}
}

func sampleSnapshot() model.Snapshot {
	return model.Snapshot{
		RepoName: "coordination",
		Branch:   "jj",
		Items: []model.TodoItem{
			{TodoID: "TODO-binap", Status: model.StatusOpen},
		},
		OrphanDetail: []string{"TODO/TODO-orphan.md"},
	}
}
