package lint

import (
	"fmt"
	"sort"

	"todo-report/internal/model"
)

func Run(snapshot model.Snapshot) []model.LintFinding {
	findings := append([]model.LintFinding{}, snapshot.Findings...)

	for _, item := range snapshot.Items {
		declaredStatus := snapshot.IndexStatusByID[item.TodoID]
		subtasks := snapshot.SubtasksByParent[item.TodoID]
		if len(subtasks) == 0 {
			continue
		}

		openCount := 0
		for _, subtask := range subtasks {
			if subtask.Status == model.StatusOpen {
				openCount++
			}
		}
		allSubtasksCompleted := openCount == 0

		if declaredStatus == model.StatusOpen && (item.Status == model.StatusCompleted || allSubtasksCompleted) {
			findings = append(findings, model.LintFinding{
				Severity: "warning",
				Code:     "index_open_detail_complete",
				TodoID:   item.TodoID,
				File:     item.SourceFile,
				Line:     item.Line,
				Message:  fmt.Sprintf("Index item %s remains open while its detail file appears complete.", item.TodoID),
			})
		}
		if declaredStatus == model.StatusCompleted && openCount > 0 {
			findings = append(findings, model.LintFinding{
				Severity: "warning",
				Code:     "index_done_detail_open",
				TodoID:   item.TodoID,
				File:     item.SourceFile,
				Line:     item.Line,
				Message:  fmt.Sprintf("Index item %s is checked while %d subtask(s) remain open in the detail file.", item.TodoID, openCount),
			})
		}
	}

	for _, file := range snapshot.OrphanDetail {
		findings = append(findings, model.LintFinding{
			Severity: "warning",
			Code:     "orphan_detail_file",
			File:     file,
			Line:     1,
			Message:  fmt.Sprintf("Detail file exists under %s but is not linked from %s.", snapshot.TodoRoot, snapshot.IndexFile),
		})
	}

	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Severity == findings[j].Severity {
			if findings[i].File == findings[j].File {
				return findings[i].Line < findings[j].Line
			}
			return findings[i].File < findings[j].File
		}
		return findings[i].Severity < findings[j].Severity
	})
	return findings
}
