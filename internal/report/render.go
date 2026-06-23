package report

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func RenderHealth(snapshot model.Snapshot, ages []model.AgeRecord, findings []model.LintFinding, driftResult *model.DriftResult, format string) (string, error) {
	type health struct {
		Repo              string              `json:"repo"`
		Branch            string              `json:"branch"`
		OpenTODOs         int                 `json:"open_todos"`
		CompletedTODOs    int                 `json:"completed_todos"`
		OldestOpen        *model.AgeRecord    `json:"oldest_open,omitempty"`
		LintErrors        int                 `json:"lint_errors"`
		LintWarnings      int                 `json:"lint_warnings"`
		Drift             *model.DriftResult  `json:"drift,omitempty"`
		Findings          []model.LintFinding `json:"findings"`
		Age               []model.AgeRecord   `json:"age"`
		OrphanDetailFiles []string            `json:"orphan_detail_files"`
	}

	openCount := 0
	completedCount := 0
	for _, item := range snapshot.Items {
		if item.Status == model.StatusCompleted {
			completedCount++
		} else {
			openCount++
		}
	}
	var oldest *model.AgeRecord
	for i := range ages {
		if ages[i].Todo.Status == model.StatusCompleted {
			continue
		}
		oldest = &ages[i]
		break
	}
	errorsCount, warningsCount := countFindings(findings)
	data := health{
		Repo:              snapshot.RepoName,
		Branch:            snapshot.Branch,
		OpenTODOs:         openCount,
		CompletedTODOs:    completedCount,
		OldestOpen:        oldest,
		LintErrors:        errorsCount,
		LintWarnings:      warningsCount,
		Drift:             driftResult,
		Findings:          findings,
		Age:               ages,
		OrphanDetailFiles: snapshot.OrphanDetail,
	}

	switch format {
	case "json":
		return marshal(data)
	case "tsv":
		var b strings.Builder
		b.WriteString("key\tvalue\n")
		fmt.Fprintf(&b, "repo\t%s\n", data.Repo)
		fmt.Fprintf(&b, "branch\t%s\n", data.Branch)
		fmt.Fprintf(&b, "open_todos\t%d\n", data.OpenTODOs)
		fmt.Fprintf(&b, "completed_todos\t%d\n", data.CompletedTODOs)
		if oldest != nil {
			fmt.Fprintf(&b, "oldest_open\t%s\n", oldest.Todo.TodoID)
			fmt.Fprintf(&b, "oldest_open_age_days\t%d\n", oldest.AgeDays)
		}
		fmt.Fprintf(&b, "lint_errors\t%d\n", data.LintErrors)
		fmt.Fprintf(&b, "lint_warnings\t%d\n", data.LintWarnings)
		if driftResult != nil {
			fmt.Fprintf(&b, "drift_items\t%d\n", driftResult.TotalDifferenceRows)
		}
		return b.String(), nil
	case "markdown":
		var b strings.Builder
		b.WriteString("## Health Report\n\n")
		fmt.Fprintf(&b, "- Repo: `%s`\n", data.Repo)
		fmt.Fprintf(&b, "- Branch: `%s`\n", data.Branch)
		fmt.Fprintf(&b, "- Open TODOs: %d\n", data.OpenTODOs)
		fmt.Fprintf(&b, "- Completed TODOs: %d\n", data.CompletedTODOs)
		if oldest != nil {
			fmt.Fprintf(&b, "- Oldest open TODO: `%s` (%d days)\n", oldest.Todo.TodoID, oldest.AgeDays)
		}
		fmt.Fprintf(&b, "- Lint errors: %d\n", data.LintErrors)
		fmt.Fprintf(&b, "- Lint warnings: %d\n\n", data.LintWarnings)
		if driftResult != nil {
			driftMD, _ := RenderDrift(*driftResult, "markdown")
			b.WriteString(driftMD)
		}
		lintMD, _ := RenderLint(snapshot, findings, "markdown")
		b.WriteString(lintMD)
		return b.String(), nil
	case "text":
		var b strings.Builder
		fmt.Fprintf(&b, "Repo: %s\n", data.Repo)
		fmt.Fprintf(&b, "Branch: %s\n", data.Branch)
		fmt.Fprintf(&b, "Open TODOs: %d\n", data.OpenTODOs)
		fmt.Fprintf(&b, "Completed TODOs: %d\n", data.CompletedTODOs)
		if oldest != nil {
			fmt.Fprintf(&b, "Oldest open TODO: %s, %d days\n", oldest.Todo.TodoID, oldest.AgeDays)
		}
		fmt.Fprintf(&b, "Lint errors: %d\n", data.LintErrors)
		fmt.Fprintf(&b, "Lint warnings: %d\n", data.LintWarnings)
		if driftResult != nil {
			fmt.Fprintf(&b, "Branch drift items: %d\n", driftResult.TotalDifferenceRows)
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

func countFindings(findings []model.LintFinding) (errorsCount, warningsCount int) {
	for _, finding := range findings {
		switch finding.Severity {
		case "error":
			errorsCount++
		case "warning":
			warningsCount++
		}
	}
	return errorsCount, warningsCount
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
