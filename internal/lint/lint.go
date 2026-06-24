package lint

import (
	"fmt"
	"sort"

	"todo-report/internal/model"
)

func Run(snapshot model.Snapshot) []model.LintFinding {
	findings := append([]model.LintFinding{}, snapshot.Findings...)

	for _, item := range snapshot.Items {
		subtasks := snapshot.SubtasksByParent[item.TodoID]
		if item.Status == model.StatusCompleted {
			for _, subtask := range subtasks {
				if subtask.Status == model.StatusOpen {
					findings = append(findings, model.LintFinding{
						Severity: "warning",
						Code:     "checked_parent_open_subtask",
						TodoID:   item.TodoID,
						File:     subtask.SourceFile,
						Line:     subtask.Line,
						Message:  fmt.Sprintf("Parent %s is checked while subtask %s remains open.", item.TodoID, subtask.SubtaskID),
					})
				}
			}
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
