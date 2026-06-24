package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"

	"todo-report/internal/model"
)

func RenderAge(records []model.AgeRecord, format string) (string, error) {
	switch format {
	case "json":
		return marshal(records)
	case "tsv":
		var b strings.Builder
		b.WriteString("todo_id\tstatus\tage_days\tfirst_seen\ttitle\tsource_file\tdetail_file\n")
		for _, record := range records {
			fmt.Fprintf(&b, "%s\t%s\t%d\t%s\t%s\t%s\t%s\n",
				record.Todo.TodoID,
				record.Todo.Status,
				record.AgeDays,
				record.FirstSeen.Format("2006-01-02"),
				record.Todo.Title,
				record.Todo.SourceFile,
				record.Todo.DetailFile,
			)
		}
		return b.String(), nil
	case "markdown":
		var b strings.Builder
		b.WriteString("## Age Report\n\n")
		for _, record := range records {
			fmt.Fprintf(&b, "- [ ] `%s` - %s (%d days; first seen %s)\n",
				record.Todo.TodoID,
				record.Todo.Title,
				record.AgeDays,
				record.FirstSeen.Format("2006-01-02"),
			)
		}
		return b.String(), nil
	case "text":
		var b strings.Builder
		for _, record := range records {
			fmt.Fprintf(&b, "%s\t%d days\t%s\t%s\n",
				record.Todo.TodoID,
				record.AgeDays,
				record.Todo.SourceFile,
				record.Todo.Title,
			)
		}
		return b.String(), nil
	default:
		return "", fmt.Errorf("unsupported format %q", format)
	}
}

func RenderLint(snapshot model.Snapshot, findings []model.LintFinding, format string) (string, error) {
	switch format {
	case "json":
		return marshal(struct {
			Snapshot model.Snapshot      `json:"snapshot"`
			Findings []model.LintFinding `json:"findings"`
		}{Snapshot: snapshot, Findings: findings})
	case "tsv":
		var b strings.Builder
		b.WriteString("severity\tcode\ttodo_id\tfile\tline\tmessage\n")
		for _, finding := range findings {
			fmt.Fprintf(&b, "%s\t%s\t%s\t%s\t%d\t%s\n",
				finding.Severity, finding.Code, finding.TodoID, finding.File, finding.Line, finding.Message)
		}
		return b.String(), nil
	case "markdown":
		var b strings.Builder
		b.WriteString("## Lint Report\n\n")
		if len(findings) == 0 {
			b.WriteString("- [x] No lint findings\n")
			return b.String(), nil
		}
		for _, finding := range findings {
			fmt.Fprintf(&b, "- [ ] `%s` `%s:%d`", finding.Code, finding.File, finding.Line)
			if finding.TodoID != "" {
				fmt.Fprintf(&b, " `%s`", finding.TodoID)
			}
			fmt.Fprintf(&b, " - %s\n", finding.Message)
		}
		return b.String(), nil
	case "text":
		var b strings.Builder
		for _, finding := range findings {
			fmt.Fprintf(&b, "%s %s %s:%d %s\n",
				strings.ToUpper(finding.Severity), finding.Code, finding.File, finding.Line, finding.Message)
		}
		return b.String(), nil
	default:
		return "", fmt.Errorf("unsupported format %q", format)
	}
}

func RenderDrift(result model.DriftResult, format string) (string, error) {
	switch format {
	case "json":
		return marshal(result)
	case "tsv":
		var b strings.Builder
		b.WriteString("kind\tvalue\tdetails\n")
		writeListTSV(&b, "only_in_"+result.BranchA, result.OnlyInA)
		writeListTSV(&b, "only_in_"+result.BranchB, result.OnlyInB)
		writeListTSV(&b, "completed_only_in_"+result.BranchA, result.CompletedOnlyInA)
		writeListTSV(&b, "completed_only_in_"+result.BranchB, result.CompletedOnlyInB)
		writeListTSV(&b, "detail_only_in_"+result.BranchA, result.DetailOnlyInA)
		writeListTSV(&b, "detail_only_in_"+result.BranchB, result.DetailOnlyInB)
		writeListTSV(&b, "subtask_only_in_"+result.BranchA, result.SubtaskOnlyInA)
		writeListTSV(&b, "subtask_only_in_"+result.BranchB, result.SubtaskOnlyInB)
		writeListTSV(&b, "subtask_completed_only_in_"+result.BranchA, result.SubtaskCompletedA)
		writeListTSV(&b, "subtask_completed_only_in_"+result.BranchB, result.SubtaskCompletedB)
		for _, diff := range result.OtherDifferences {
			label := diff.Kind
			value := diff.TodoID
			if value == "" {
				value = diff.Target
			}
			fmt.Fprintf(&b, "%s\t%s\t%s\n", label, value, diff.Details)
		}
		return b.String(), nil
	case "markdown":
		var b strings.Builder
		b.WriteString("## Drift Report\n\n")
		fmt.Fprintf(&b, "- Branch A: `%s`\n", result.BranchA)
		fmt.Fprintf(&b, "- Branch B: `%s`\n\n", result.BranchB)
		writeMarkdownSection(&b, "Only in "+result.BranchA, result.OnlyInA)
		writeMarkdownSection(&b, "Only in "+result.BranchB, result.OnlyInB)
		writeMarkdownSection(&b, "Completed only in "+result.BranchA, result.CompletedOnlyInA)
		writeMarkdownSection(&b, "Completed only in "+result.BranchB, result.CompletedOnlyInB)
		writeMarkdownSection(&b, "Detail files only in "+result.BranchA, result.DetailOnlyInA)
		writeMarkdownSection(&b, "Detail files only in "+result.BranchB, result.DetailOnlyInB)
		writeMarkdownSection(&b, "Subtasks only in "+result.BranchA, result.SubtaskOnlyInA)
		writeMarkdownSection(&b, "Subtasks only in "+result.BranchB, result.SubtaskOnlyInB)
		writeMarkdownSection(&b, "Subtasks completed only in "+result.BranchA, result.SubtaskCompletedA)
		writeMarkdownSection(&b, "Subtasks completed only in "+result.BranchB, result.SubtaskCompletedB)
		if len(result.OtherDifferences) > 0 {
			b.WriteString("### Other Differences\n\n")
			for _, diff := range result.OtherDifferences {
				label := diff.TodoID
				if label == "" {
					label = diff.Target
				}
				fmt.Fprintf(&b, "- [ ] `%s` `%s` - %s\n", diff.Kind, label, diff.Details)
			}
			b.WriteString("\n")
		}
		return b.String(), nil
	case "text":
		var b strings.Builder
		writeTextSection(&b, "Only in "+result.BranchA, result.OnlyInA)
		writeTextSection(&b, "Only in "+result.BranchB, result.OnlyInB)
		writeTextSection(&b, "Completed only in "+result.BranchA, result.CompletedOnlyInA)
		writeTextSection(&b, "Completed only in "+result.BranchB, result.CompletedOnlyInB)
		writeTextSection(&b, "Detail files only in "+result.BranchA, result.DetailOnlyInA)
		writeTextSection(&b, "Detail files only in "+result.BranchB, result.DetailOnlyInB)
		writeTextSection(&b, "Subtasks only in "+result.BranchA, result.SubtaskOnlyInA)
		writeTextSection(&b, "Subtasks only in "+result.BranchB, result.SubtaskOnlyInB)
		writeTextSection(&b, "Subtasks completed only in "+result.BranchA, result.SubtaskCompletedA)
		writeTextSection(&b, "Subtasks completed only in "+result.BranchB, result.SubtaskCompletedB)
		for _, diff := range result.OtherDifferences {
			label := diff.TodoID
			if label == "" {
				label = diff.Target
			}
			fmt.Fprintf(&b, "%s: %s (%s)\n", diff.Kind, label, diff.Details)
		}
		return b.String(), nil
	default:
		return "", fmt.Errorf("unsupported format %q", format)
	}
}

func RenderHealth(report model.HealthReport, format string) (string, error) {
	switch format {
	case "json":
		return marshal(report)
	case "tsv":
		var b strings.Builder
		b.WriteString("section\tkey\tvalue\n")
		fmt.Fprintf(&b, "summary\trepo\t%s\n", report.Repo)
		fmt.Fprintf(&b, "summary\tbranch\t%s\n", report.Branch)
		fmt.Fprintf(&b, "summary\tindex_file\t%s\n", report.IndexFile)
		fmt.Fprintf(&b, "summary\tstatus\t%s\n", report.Status)
		fmt.Fprintf(&b, "summary\topen_todos\t%d\n", report.OpenTODOs)
		fmt.Fprintf(&b, "summary\tcompleted_todos\t%d\n", report.CompletedTODOs)
		if report.OldestOpen != nil {
			fmt.Fprintf(&b, "summary\toldest_open\t%s\n", report.OldestOpen.Todo.TodoID)
			fmt.Fprintf(&b, "summary\toldest_open_age_days\t%d\n", report.OldestOpen.AgeDays)
		}
		fmt.Fprintf(&b, "summary\tlint_errors\t%d\n", report.LintErrors)
		fmt.Fprintf(&b, "summary\tlint_warnings\t%d\n", report.LintWarnings)
		for _, summary := range report.FindingSummary {
			fmt.Fprintf(&b, "finding_summary\t%s/%s\t%d\n", summary.Severity, summary.Code, summary.Count)
		}
		for _, record := range report.OldestOpenItems {
			fmt.Fprintf(&b, "oldest_open\t%s\t%d\n", record.Todo.TodoID, record.AgeDays)
		}
		if report.Drift != nil {
			fmt.Fprintf(&b, "summary\tdrift_items\t%d\n", report.Drift.TotalDifferenceRows)
		}
		return b.String(), nil
	case "markdown":
		var b strings.Builder
		b.WriteString("## Health Report\n\n")
		fmt.Fprintf(&b, "- Repo: `%s`\n", report.Repo)
		fmt.Fprintf(&b, "- Branch: `%s`\n", report.Branch)
		fmt.Fprintf(&b, "- Index: `%s`\n", report.IndexFile)
		fmt.Fprintf(&b, "- Status: `%s`\n", report.Status)
		fmt.Fprintf(&b, "- Open TODOs: %d\n", report.OpenTODOs)
		fmt.Fprintf(&b, "- Completed TODOs: %d\n", report.CompletedTODOs)
		if report.OldestOpen != nil {
			fmt.Fprintf(&b, "- Oldest open TODO: `%s` (%d days)\n", report.OldestOpen.Todo.TodoID, report.OldestOpen.AgeDays)
		}
		fmt.Fprintf(&b, "- Lint errors: %d\n", report.LintErrors)
		fmt.Fprintf(&b, "- Lint warnings: %d\n\n", report.LintWarnings)

		if len(report.OldestOpenItems) > 0 {
			b.WriteString("### Oldest Open TODOs\n\n")
			for _, record := range report.OldestOpenItems {
				fmt.Fprintf(&b, "- [ ] `%s` - %s (%d days)\n", record.Todo.TodoID, record.Todo.Title, record.AgeDays)
			}
			b.WriteString("\n")
		}

		if len(report.FindingSummary) > 0 {
			b.WriteString("### Finding Summary\n\n")
			for _, summary := range report.FindingSummary {
				fmt.Fprintf(&b, "- [ ] `%s` `%s` - %d\n", summary.Severity, summary.Code, summary.Count)
			}
			b.WriteString("\n")
		}

		if report.Drift != nil {
			b.WriteString("### Drift Summary\n\n")
			fmt.Fprintf(&b, "- Difference rows: %d\n\n", report.Drift.TotalDifferenceRows)
			driftMD, _ := RenderDrift(*report.Drift, "markdown")
			b.WriteString(driftMD)
		}
		lintMD, _ := RenderLint(model.Snapshot{}, report.Findings, "markdown")
		b.WriteString(lintMD)
		return b.String(), nil
	case "text":
		var b strings.Builder
		fmt.Fprintf(&b, "Repo: %s\n", report.Repo)
		fmt.Fprintf(&b, "Branch: %s\n", report.Branch)
		fmt.Fprintf(&b, "Index: %s\n", report.IndexFile)
		fmt.Fprintf(&b, "Status: %s\n", strings.ToUpper(report.Status))
		fmt.Fprintf(&b, "Open TODOs: %d\n", report.OpenTODOs)
		fmt.Fprintf(&b, "Completed TODOs: %d\n", report.CompletedTODOs)
		if report.OldestOpen != nil {
			fmt.Fprintf(&b, "Oldest open TODO: %s, %d days\n", report.OldestOpen.Todo.TodoID, report.OldestOpen.AgeDays)
		}
		fmt.Fprintf(&b, "Lint errors: %d\n", report.LintErrors)
		fmt.Fprintf(&b, "Lint warnings: %d\n", report.LintWarnings)
		if report.Drift != nil {
			fmt.Fprintf(&b, "Branch drift items: %d\n", report.Drift.TotalDifferenceRows)
		}
		if len(report.FindingSummary) > 0 {
			b.WriteString("\nTop findings:\n")
			for _, summary := range report.FindingSummary {
				fmt.Fprintf(&b, "  %s %s (%d)\n", strings.ToUpper(summary.Severity), summary.Code, summary.Count)
			}
		}
		if len(report.OldestOpenItems) > 0 {
			b.WriteString("\nOldest open TODOs:\n")
			for _, record := range report.OldestOpenItems {
				fmt.Fprintf(&b, "  %s\t%d days\t%s\n", record.Todo.TodoID, record.AgeDays, record.Todo.Title)
			}
		}
		return b.String(), nil
	default:
		return "", fmt.Errorf("unsupported format %q", format)
	}
}

func RenderMultiHealth(report model.MultiHealthReport, format string) (string, error) {
	switch format {
	case "json":
		return marshal(report)
	case "tsv":
		var b strings.Builder
		b.WriteString("scope\tindex_file\tstatus\topen_todos\tcompleted_todos\tlint_errors\tlint_warnings\tdrift_items\n")
		fmt.Fprintf(&b, "summary\t(all)\t%s\t%d\t%d\t%d\t%d\t%d\n", report.Status, report.OpenTODOs, report.CompletedTODOs, report.LintErrors, report.LintWarnings, report.DriftItems)
		for _, entry := range report.Reports {
			driftItems := 0
			if entry.Drift != nil {
				driftItems = entry.Drift.TotalDifferenceRows
			}
			fmt.Fprintf(&b, "index\t%s\t%s\t%d\t%d\t%d\t%d\t%d\n", entry.IndexFile, entry.Status, entry.OpenTODOs, entry.CompletedTODOs, entry.LintErrors, entry.LintWarnings, driftItems)
		}
		for _, index := range report.IndexesOnlyInBranch {
			fmt.Fprintf(&b, "only_in_%s\t%s\twarning\t0\t0\t0\t0\t0\n", report.Branch, index)
		}
		for _, index := range report.IndexesOnlyInCompare {
			fmt.Fprintf(&b, "only_in_%s\t%s\twarning\t0\t0\t0\t0\t0\n", report.CompareBranch, index)
		}
		return b.String(), nil
	case "markdown":
		var b strings.Builder
		b.WriteString("## Multi-Index Health Report\n\n")
		fmt.Fprintf(&b, "- Repo: `%s`\n", report.Repo)
		fmt.Fprintf(&b, "- Branch: `%s`\n", report.Branch)
		if report.CompareBranch != "" {
			fmt.Fprintf(&b, "- Compare branch: `%s`\n", report.CompareBranch)
		}
		fmt.Fprintf(&b, "- Status: `%s`\n", report.Status)
		fmt.Fprintf(&b, "- Discovered indexes: %d\n", len(report.IndexFiles))
		fmt.Fprintf(&b, "- Open TODOs: %d\n", report.OpenTODOs)
		fmt.Fprintf(&b, "- Completed TODOs: %d\n", report.CompletedTODOs)
		fmt.Fprintf(&b, "- Lint errors: %d\n", report.LintErrors)
		fmt.Fprintf(&b, "- Lint warnings: %d\n", report.LintWarnings)
		if report.CompareBranch != "" {
			fmt.Fprintf(&b, "- Repo-wide drift rows: %d\n", report.DriftItems)
		}
		b.WriteString("\n")
		b.WriteString("### Index Summaries\n\n")
		for _, entry := range report.Reports {
			if entry.Drift != nil {
				fmt.Fprintf(&b, "- [ ] `%s` - status `%s`, open %d, completed %d, lint errors %d, lint warnings %d, drift rows %d\n",
					entry.IndexFile, entry.Status, entry.OpenTODOs, entry.CompletedTODOs, entry.LintErrors, entry.LintWarnings, entry.Drift.TotalDifferenceRows)
				continue
			}
			fmt.Fprintf(&b, "- [ ] `%s` - status `%s`, open %d, completed %d, lint errors %d, lint warnings %d\n",
				entry.IndexFile, entry.Status, entry.OpenTODOs, entry.CompletedTODOs, entry.LintErrors, entry.LintWarnings)
		}
		b.WriteString("\n")
		if len(report.IndexesOnlyInBranch) > 0 {
			fmt.Fprintf(&b, "### Indexes Only in %s\n\n", report.Branch)
			for _, index := range report.IndexesOnlyInBranch {
				fmt.Fprintf(&b, "- [ ] `%s`\n", index)
			}
			b.WriteString("\n")
		}
		if len(report.IndexesOnlyInCompare) > 0 {
			fmt.Fprintf(&b, "### Indexes Only in %s\n\n", report.CompareBranch)
			for _, index := range report.IndexesOnlyInCompare {
				fmt.Fprintf(&b, "- [ ] `%s`\n", index)
			}
			b.WriteString("\n")
		}
		return b.String(), nil
	case "text":
		var b strings.Builder
		fmt.Fprintf(&b, "Repo: %s\n", report.Repo)
		fmt.Fprintf(&b, "Branch: %s\n", report.Branch)
		if report.CompareBranch != "" {
			fmt.Fprintf(&b, "Compare branch: %s\n", report.CompareBranch)
		}
		fmt.Fprintf(&b, "Status: %s\n", strings.ToUpper(report.Status))
		fmt.Fprintf(&b, "Discovered indexes: %d\n", len(report.IndexFiles))
		fmt.Fprintf(&b, "Open TODOs: %d\n", report.OpenTODOs)
		fmt.Fprintf(&b, "Completed TODOs: %d\n", report.CompletedTODOs)
		fmt.Fprintf(&b, "Lint errors: %d\n", report.LintErrors)
		fmt.Fprintf(&b, "Lint warnings: %d\n", report.LintWarnings)
		if report.CompareBranch != "" {
			fmt.Fprintf(&b, "Repo-wide drift rows: %d\n", report.DriftItems)
		}
		b.WriteString("\nIndex summaries:\n")
		for _, entry := range report.Reports {
			driftSuffix := ""
			if entry.Drift != nil {
				driftSuffix = fmt.Sprintf("\tdrift=%d", entry.Drift.TotalDifferenceRows)
			}
			fmt.Fprintf(&b, "  %s\t%s\topen=%d\tcompleted=%d\terrors=%d\twarnings=%d%s\n",
				entry.IndexFile, strings.ToUpper(entry.Status), entry.OpenTODOs, entry.CompletedTODOs, entry.LintErrors, entry.LintWarnings, driftSuffix)
		}
		if len(report.IndexesOnlyInBranch) > 0 {
			fmt.Fprintf(&b, "\nIndexes only in %s:\n", report.Branch)
			for _, index := range report.IndexesOnlyInBranch {
				fmt.Fprintf(&b, "  %s\n", index)
			}
		}
		if len(report.IndexesOnlyInCompare) > 0 {
			fmt.Fprintf(&b, "\nIndexes only in %s:\n", report.CompareBranch)
			for _, index := range report.IndexesOnlyInCompare {
				fmt.Fprintf(&b, "  %s\n", index)
			}
		}
		return b.String(), nil
	default:
		return "", fmt.Errorf("unsupported format %q", format)
	}
}

func RenderIndexes(indexes []string, format string) (string, error) {
	switch format {
	case "json":
		return marshal(struct {
			Indexes []string `json:"indexes"`
		}{Indexes: indexes})
	case "tsv":
		var b strings.Builder
		b.WriteString("index_file\ttodo_root\n")
		for _, index := range indexes {
			fmt.Fprintf(&b, "%s\t%s\n", index, path.Dir(index))
		}
		return b.String(), nil
	case "markdown":
		var b strings.Builder
		b.WriteString("## TODO Indexes\n\n")
		if len(indexes) == 0 {
			b.WriteString("- [x] None found\n")
			return b.String(), nil
		}
		for _, index := range indexes {
			fmt.Fprintf(&b, "- [ ] `%s`\n", index)
		}
		return b.String(), nil
	case "text":
		var b strings.Builder
		for _, index := range indexes {
			b.WriteString(index)
			b.WriteByte('\n')
		}
		return b.String(), nil
	default:
		return "", fmt.Errorf("unsupported format %q", format)
	}
}

func RenderFleetHealth(report model.FleetHealthReport, format string) (string, error) {
	switch format {
	case "json":
		return marshal(report)
	case "tsv":
		var b strings.Builder
		b.WriteString("scope\trepo\trepo_path\tstatus\tindex_mode\tindex_count\topen_todos\tcompleted_todos\tlint_errors\tlint_warnings\tdrift_items\terror\n")
		fmt.Fprintf(&b, "summary\t(all)\t%s\t%s\t-\t%d\t%d\t%d\t%d\t%d\t%d\t%d\n",
			report.RepoListFile, report.Status, report.RepoCount, report.OpenTODOs, report.CompletedTODOs, report.LintErrors, report.LintWarnings, report.DriftItems, report.ErrorCount)
		for _, entry := range report.Entries {
			fmt.Fprintf(&b, "repo\t%s\t%s\t%s\t%s\t%d\t%d\t%d\t%d\t%d\t%d\t%s\n",
				entry.Repo, entry.RepoPath, entry.Status, entry.IndexMode, entry.IndexCount, entry.OpenTODOs, entry.CompletedTODOs, entry.LintErrors, entry.LintWarnings, entry.DriftItems, cleanTSV(entry.Error))
		}
		return b.String(), nil
	case "markdown":
		var b strings.Builder
		b.WriteString("## Fleet Health Report\n\n")
		fmt.Fprintf(&b, "- Branch: `%s`\n", report.Branch)
		if report.CompareBranch != "" {
			fmt.Fprintf(&b, "- Compare branch: `%s`\n", report.CompareBranch)
		}
		fmt.Fprintf(&b, "- Repo list: `%s`\n", report.RepoListFile)
		fmt.Fprintf(&b, "- Status: `%s`\n", report.Status)
		fmt.Fprintf(&b, "- Repos: %d\n", report.RepoCount)
		fmt.Fprintf(&b, "- Successful repos: %d\n", report.SuccessCount)
		fmt.Fprintf(&b, "- Repo errors: %d\n", report.ErrorCount)
		fmt.Fprintf(&b, "- Open TODOs: %d\n", report.OpenTODOs)
		fmt.Fprintf(&b, "- Completed TODOs: %d\n", report.CompletedTODOs)
		fmt.Fprintf(&b, "- Lint errors: %d\n", report.LintErrors)
		fmt.Fprintf(&b, "- Lint warnings: %d\n", report.LintWarnings)
		if report.CompareBranch != "" {
			fmt.Fprintf(&b, "- Fleet drift rows: %d\n", report.DriftItems)
		}
		b.WriteString("\n")

		attentionRepos := countFleetAttentionRepos(report.Entries)
		b.WriteString("### Fleet Summary\n\n")
		fmt.Fprintf(&b, "- [ ] Fleet status: `%s`\n", report.Status)
		fmt.Fprintf(&b, "- [ ] Repos needing attention: %d\n", attentionRepos)
		if report.ErrorCount > 0 {
			fmt.Fprintf(&b, "- [ ] Repo load failures: %d\n", report.ErrorCount)
		} else {
			b.WriteString("- [x] Repo load failures: 0\n")
		}
		if report.LintErrors > 0 {
			fmt.Fprintf(&b, "- [ ] Fleet lint errors: %d\n", report.LintErrors)
		} else {
			b.WriteString("- [x] Fleet lint errors: 0\n")
		}
		if report.CompareBranch != "" {
			if report.DriftItems > 0 {
				fmt.Fprintf(&b, "- [ ] Fleet drift rows: %d\n", report.DriftItems)
			} else {
				b.WriteString("- [x] Fleet drift rows: 0\n")
			}
		}
		b.WriteString("\n")

		writeFleetRepoFailuresMarkdown(&b, report.Entries)
		writeFleetAttentionMarkdown(&b, report.Entries)
		writeFleetRepoSectionsMarkdown(&b, report.Entries)
		return b.String(), nil
	case "text":
		var b strings.Builder
		fmt.Fprintf(&b, "Branch: %s\n", report.Branch)
		if report.CompareBranch != "" {
			fmt.Fprintf(&b, "Compare branch: %s\n", report.CompareBranch)
		}
		fmt.Fprintf(&b, "Repo list: %s\n", report.RepoListFile)
		fmt.Fprintf(&b, "Status: %s\n", strings.ToUpper(report.Status))
		fmt.Fprintf(&b, "Repos: %d\n", report.RepoCount)
		fmt.Fprintf(&b, "Successful repos: %d\n", report.SuccessCount)
		fmt.Fprintf(&b, "Repo errors: %d\n", report.ErrorCount)
		fmt.Fprintf(&b, "Open TODOs: %d\n", report.OpenTODOs)
		fmt.Fprintf(&b, "Completed TODOs: %d\n", report.CompletedTODOs)
		fmt.Fprintf(&b, "Lint errors: %d\n", report.LintErrors)
		fmt.Fprintf(&b, "Lint warnings: %d\n", report.LintWarnings)
		if report.CompareBranch != "" {
			fmt.Fprintf(&b, "Fleet drift rows: %d\n", report.DriftItems)
		}
		b.WriteString("\nRepos:\n")
		for _, entry := range report.Entries {
			if entry.Error != "" {
				fmt.Fprintf(&b, "  %s\tERROR\t%s\n", entry.RepoPath, entry.Error)
				continue
			}
			fmt.Fprintf(&b, "  %s\t%s\tmode=%s\tindexes=%d\topen=%d\tcompleted=%d\terrors=%d\twarnings=%d",
				entry.Repo, strings.ToUpper(entry.Status), entry.IndexMode, entry.IndexCount, entry.OpenTODOs, entry.CompletedTODOs, entry.LintErrors, entry.LintWarnings)
			if report.CompareBranch != "" {
				fmt.Fprintf(&b, "\tdrift=%d", entry.DriftItems)
			}
			b.WriteByte('\n')
		}
		return b.String(), nil
	default:
		return "", fmt.Errorf("unsupported format %q", format)
	}
}

func marshal(v any) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data) + "\n", nil
}

func writeTextSection(b *strings.Builder, title string, values []string) {
	fmt.Fprintf(b, "%s:\n", title)
	if len(values) == 0 {
		b.WriteString("  (none)\n")
		return
	}
	for _, value := range values {
		fmt.Fprintf(b, "  %s\n", value)
	}
}

func writeMarkdownSection(b *strings.Builder, title string, values []string) {
	b.WriteString("### " + title + "\n\n")
	if len(values) == 0 {
		b.WriteString("- [x] None\n\n")
		return
	}
	for _, value := range values {
		fmt.Fprintf(b, "- [ ] `%s`\n", value)
	}
	b.WriteString("\n")
}

func writeListTSV(b *strings.Builder, kind string, values []string) {
	for _, value := range values {
		fmt.Fprintf(b, "%s\t%s\t\n", kind, value)
	}
}

func writeFleetRepoFailuresMarkdown(b *strings.Builder, entries []model.FleetHealthEntry) {
	var failed []model.FleetHealthEntry
	for _, entry := range entries {
		if entry.Error != "" {
			failed = append(failed, entry)
		}
	}
	if len(failed) == 0 {
		return
	}

	b.WriteString("### Repo Failures\n\n")
	for _, entry := range failed {
		fmt.Fprintf(b, "- [ ] `%s` - %s\n", entry.RepoPath, entry.Error)
	}
	b.WriteString("\n")
}

func writeFleetAttentionMarkdown(b *strings.Builder, entries []model.FleetHealthEntry) {
	var attention []model.FleetHealthEntry
	for _, entry := range entries {
		if entry.Error != "" || entry.Status == "clean" {
			continue
		}
		attention = append(attention, entry)
	}
	if len(attention) == 0 {
		b.WriteString("### Repos Needing Attention\n\n")
		b.WriteString("- [x] None\n\n")
		return
	}

	b.WriteString("### Repos Needing Attention\n\n")
	for _, entry := range attention {
		line := fmt.Sprintf("- [ ] `%s` - status `%s`, open %d, completed %d, lint errors %d, lint warnings %d",
			entry.Repo, entry.Status, entry.OpenTODOs, entry.CompletedTODOs, entry.LintErrors, entry.LintWarnings)
		if entry.DriftItems > 0 {
			line += fmt.Sprintf(", drift rows %d", entry.DriftItems)
		}
		b.WriteString(line + "\n")
	}
	b.WriteString("\n")
}

func writeFleetRepoSectionsMarkdown(b *strings.Builder, entries []model.FleetHealthEntry) {
	b.WriteString("### Repo Details\n\n")
	for _, entry := range entries {
		label := entry.Repo
		if label == "" {
			label = entry.RepoPath
		}
		fmt.Fprintf(b, "#### `%s`\n\n", label)
		fmt.Fprintf(b, "- Path: `%s`\n", entry.RepoPath)
		fmt.Fprintf(b, "- Status: `%s`\n", entry.Status)
		fmt.Fprintf(b, "- Mode: `%s`\n", entry.IndexMode)
		if entry.Error != "" {
			fmt.Fprintf(b, "- Error: %s\n\n", entry.Error)
			continue
		}
		fmt.Fprintf(b, "- Indexes: %d\n", entry.IndexCount)
		fmt.Fprintf(b, "- Open TODOs: %d\n", entry.OpenTODOs)
		fmt.Fprintf(b, "- Completed TODOs: %d\n", entry.CompletedTODOs)
		fmt.Fprintf(b, "- Lint errors: %d\n", entry.LintErrors)
		fmt.Fprintf(b, "- Lint warnings: %d\n", entry.LintWarnings)
		if entry.DriftItems > 0 {
			fmt.Fprintf(b, "- Drift rows: %d\n", entry.DriftItems)
		}
		b.WriteString("\n")

		if entry.Health != nil {
			writeFleetSingleHealthHighlightsMarkdown(b, *entry.Health)
		}
		if entry.MultiHealth != nil {
			writeFleetMultiHealthHighlightsMarkdown(b, *entry.MultiHealth)
		}
	}
}

func writeFleetSingleHealthHighlightsMarkdown(b *strings.Builder, report model.HealthReport) {
	if len(report.FindingSummary) > 0 {
		b.WriteString("Top findings:\n")
		for _, summary := range report.FindingSummary {
			fmt.Fprintf(b, "- [ ] `%s` `%s` - %d\n", summary.Severity, summary.Code, summary.Count)
		}
	}
	if len(report.OldestOpenItems) > 0 {
		b.WriteString("Oldest open TODOs:\n")
		for _, record := range report.OldestOpenItems {
			fmt.Fprintf(b, "- [ ] `%s` - %s (%d days)\n", record.Todo.TodoID, record.Todo.Title, record.AgeDays)
		}
	}
	if report.Drift != nil && report.Drift.TotalDifferenceRows > 0 {
		fmt.Fprintf(b, "- [ ] Drift rows on `%s`: %d\n", report.IndexFile, report.Drift.TotalDifferenceRows)
	}
	if len(report.FindingSummary) > 0 || len(report.OldestOpenItems) > 0 || (report.Drift != nil && report.Drift.TotalDifferenceRows > 0) {
		b.WriteString("\n")
	}
}

func writeFleetMultiHealthHighlightsMarkdown(b *strings.Builder, report model.MultiHealthReport) {
	var highlighted []model.HealthReport
	for _, child := range report.Reports {
		if child.Status != "clean" || (child.Drift != nil && child.Drift.TotalDifferenceRows > 0) {
			highlighted = append(highlighted, child)
		}
	}
	if len(highlighted) > 0 {
		b.WriteString("Index highlights:\n")
		for _, child := range highlighted {
			line := fmt.Sprintf("- [ ] `%s` - status `%s`, lint errors %d, lint warnings %d",
				child.IndexFile, child.Status, child.LintErrors, child.LintWarnings)
			if child.Drift != nil && child.Drift.TotalDifferenceRows > 0 {
				line += fmt.Sprintf(", drift rows %d", child.Drift.TotalDifferenceRows)
			}
			b.WriteString(line + "\n")
		}
	}
	if len(report.IndexesOnlyInBranch) > 0 {
		fmt.Fprintf(b, "Indexes only in `%s`:\n", report.Branch)
		for _, index := range report.IndexesOnlyInBranch {
			fmt.Fprintf(b, "- [ ] `%s`\n", index)
		}
	}
	if len(report.IndexesOnlyInCompare) > 0 {
		fmt.Fprintf(b, "Indexes only in `%s`:\n", report.CompareBranch)
		for _, index := range report.IndexesOnlyInCompare {
			fmt.Fprintf(b, "- [ ] `%s`\n", index)
		}
	}
	if len(highlighted) > 0 || len(report.IndexesOnlyInBranch) > 0 || len(report.IndexesOnlyInCompare) > 0 {
		b.WriteString("\n")
	}
}

func cleanTSV(value string) string {
	value = strings.ReplaceAll(value, "\t", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	return value
}

func countFleetAttentionRepos(entries []model.FleetHealthEntry) int {
	count := 0
	for _, entry := range entries {
		if entry.Error != "" || entry.Status == "clean" {
			continue
		}
		count++
	}
	return count
}

func SortedKeys[K ~string, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func CompactLines(lines []string) string {
	var buf bytes.Buffer
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
	return buf.String()
}
