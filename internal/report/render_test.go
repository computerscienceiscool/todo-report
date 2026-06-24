package report

import (
	"strings"
	"testing"
	"time"

	"todo-report/internal/health"
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

func TestRenderDetectFormats(t *testing.T) {
	report := model.DetectReport{
		Repo:             "coordination",
		Branch:           "jj",
		IndexFile:        "TODO/TODO.md",
		Compatibility:    "compatible_with_warnings",
		IndexLayouts:     []string{"checklist_lines"},
		TopLevelIDStyles: []string{"numeric_legacy", "proquint"},
		SubtaskIDStyles:  []string{"numeric_dotted", "parent_prefixed_dotted"},
		Features:         []string{"approximate_checkboxes"},
		StyleFindings: []model.DetectFinding{
			{Code: "invalid_checkbox", Count: 2, Examples: []string{"TODO/TODO.md:12"}},
		},
		TopLevelCount:   3,
		DetailFileCount: 1,
		SubtaskCount:    4,
	}

	tests := []struct {
		format string
		want   []string
	}{
		{format: "text", want: []string{"Compatibility: COMPATIBLE_WITH_WARNINGS", "Top-level ID styles: numeric_legacy, proquint"}},
		{format: "markdown", want: []string{"## Detect Report", "### Top-level ID styles", "### Compatibility Findings"}},
		{format: "json", want: []string{`"compatibility": "compatible_with_warnings"`, `"style_findings": [`}},
		{format: "tsv", want: []string{"section\tkey\tvalue", "summary\tcompatibility\tcompatible_with_warnings", "style_finding\tinvalid_checkbox\t2"}},
	}

	for _, tc := range tests {
		t.Run(tc.format, func(t *testing.T) {
			out, err := RenderDetect(report, tc.format)
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
	report := health.Build(snapshot, ages, findings, driftResult, "main")

	tests := []struct {
		format string
		want   []string
	}{
		{format: "text", want: []string{"Repo: coordination", "Status: ERROR", "Top findings:", "Branch drift items: 4"}},
		{format: "markdown", want: []string{"## Health Report", "### Finding Summary", "## Drift Report", "## Lint Report"}},
		{format: "json", want: []string{`"open_todos": 1`, `"lint_errors": 1`, `"total_difference_rows": 4`, `"status": "error"`}},
		{format: "tsv", want: []string{"section\tkey\tvalue", "summary\trepo\tcoordination", "summary\tdrift_items\t4"}},
	}

	for _, tc := range tests {
		t.Run(tc.format, func(t *testing.T) {
			out, err := RenderHealth(report, tc.format)
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

func TestRenderMultiHealthFormats(t *testing.T) {
	reports := []model.HealthReport{
		{Repo: "coordination", Branch: "jj", IndexFile: "TODO/TODO.md", Status: "warning", OpenTODOs: 2, CompletedTODOs: 1, LintWarnings: 1, Drift: &model.DriftResult{TotalDifferenceRows: 2}},
		{Repo: "coordination", Branch: "jj", IndexFile: "protocols/wire-lab.d/TODO/TODO.md", Status: "error", OpenTODOs: 5, CompletedTODOs: 2, LintErrors: 3},
	}
	multi := health.BuildMulti("coordination", "jj", "main", reports, []string{"TODO/TODO.md"}, []string{"simulations/SIM-beta/TODO/TODO.md"})

	tests := []struct {
		format string
		want   []string
	}{
		{format: "text", want: []string{"Discovered indexes: 3", "Compare branch: main", "Repo-wide drift rows: 2", "Indexes only in jj:"}},
		{format: "markdown", want: []string{"## Multi-Index Health Report", "### Index Summaries", "### Indexes Only in jj", "`TODO/TODO.md`"}},
		{format: "json", want: []string{`"index_files": [`, `"lint_errors": 3`, `"drift_items": 2`}},
		{format: "tsv", want: []string{"scope\tindex_file\tstatus", "summary\t(all)\terror\t7\t3\t3\t1\t2", "only_in_main\tsimulations/SIM-beta/TODO/TODO.md\twarning"}},
	}

	for _, tc := range tests {
		t.Run(tc.format, func(t *testing.T) {
			out, err := RenderMultiHealth(multi, tc.format)
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

func TestRenderIndexesFormats(t *testing.T) {
	indexes := []string{"TODO/TODO.md", "protocols/wire-lab.d/TODO/TODO.md"}

	tests := []struct {
		format string
		want   []string
	}{
		{format: "text", want: []string{"TODO/TODO.md", "protocols/wire-lab.d/TODO/TODO.md"}},
		{format: "markdown", want: []string{"## TODO Indexes", "`TODO/TODO.md`"}},
		{format: "json", want: []string{`"indexes": [`}},
		{format: "tsv", want: []string{"index_file\ttodo_root", "TODO/TODO.md\tTODO"}},
	}

	for _, tc := range tests {
		t.Run(tc.format, func(t *testing.T) {
			out, err := RenderIndexes(indexes, tc.format)
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

func TestRenderFleetHealthFormats(t *testing.T) {
	report := model.FleetHealthReport{
		Branch:         "main",
		CompareBranch:  "jj",
		RepoListFile:   "/tmp/repos.txt",
		Status:         "warning",
		RepoCount:      2,
		SuccessCount:   1,
		ErrorCount:     1,
		OpenTODOs:      7,
		CompletedTODOs: 3,
		LintErrors:     0,
		LintWarnings:   2,
		DriftItems:     5,
		Entries: []model.FleetHealthEntry{
			{
				Repo:           "coordination",
				RepoPath:       "/tmp/coordination",
				Status:         "warning",
				IndexMode:      "all-indexes",
				IndexCount:     2,
				OpenTODOs:      7,
				CompletedTODOs: 3,
				LintWarnings:   2,
				DriftItems:     5,
				MultiHealth: &model.MultiHealthReport{
					Branch:               "main",
					CompareBranch:        "jj",
					IndexesOnlyInCompare: []string{"protocols/archive/TODO/TODO.md"},
					Reports: []model.HealthReport{
						{IndexFile: "TODO/TODO.md", Status: "warning", LintWarnings: 2, Drift: &model.DriftResult{TotalDifferenceRows: 5}},
					},
				},
			},
			{Repo: "broken", RepoPath: "/tmp/broken", Status: "error", Error: "open repo failed", IndexMode: "single-index"},
		},
	}

	tests := []struct {
		format string
		want   []string
	}{
		{format: "text", want: []string{"Repo list: /tmp/repos.txt", "Fleet drift rows: 5", "/tmp/broken\tERROR\topen repo failed"}},
		{format: "markdown", want: []string{"## Fleet Health Report", "### Fleet Summary", "### Repo Failures", "### Repos Needing Attention", "#### `coordination`", "Indexes only in `jj`:", "- Error: open repo failed"}},
		{format: "json", want: []string{`"repo_count": 2`, `"repo_path": "/tmp/coordination"`}},
		{format: "tsv", want: []string{"scope\trepo\trepo_path\tstatus", "summary\t(all)\t/tmp/repos.txt\twarning", "repo\tbroken\t/tmp/broken\terror"}},
	}

	for _, tc := range tests {
		t.Run(tc.format, func(t *testing.T) {
			out, err := RenderFleetHealth(report, tc.format)
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
