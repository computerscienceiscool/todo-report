package health

import (
	"sort"

	"todo-report/internal/model"
)

const topOpenLimit = 5

func Build(snapshot model.Snapshot, ages []model.AgeRecord, findings []model.LintFinding, driftResult *model.DriftResult, compareBranch string) model.HealthReport {
	openCount := 0
	completedCount := 0
	for _, item := range snapshot.Items {
		if item.Status == model.StatusCompleted {
			completedCount++
		} else {
			openCount++
		}
	}

	oldestOpenItems := oldestOpenAges(ages)
	var oldestOpen *model.AgeRecord
	if len(oldestOpenItems) > 0 {
		record := oldestOpenItems[0]
		oldestOpen = &record
	}

	errorsCount, warningsCount := countFindings(findings)
	return model.HealthReport{
		Repo:              snapshot.RepoName,
		Branch:            snapshot.Branch,
		CompareBranch:     compareBranch,
		IndexFile:         snapshot.IndexFile,
		TodoRoot:          snapshot.TodoRoot,
		Status:            deriveStatus(errorsCount, warningsCount, driftResult),
		OpenTODOs:         openCount,
		CompletedTODOs:    completedCount,
		OldestOpen:        oldestOpen,
		OldestOpenItems:   oldestOpenItems,
		LintErrors:        errorsCount,
		LintWarnings:      warningsCount,
		FindingSummary:    summarizeFindings(findings),
		Drift:             driftResult,
		Findings:          findings,
		Age:               ages,
		OrphanDetailFiles: snapshot.OrphanDetail,
	}
}

func BuildMulti(repoName, branch string, reports []model.HealthReport) model.MultiHealthReport {
	indexes := make([]string, 0, len(reports))
	result := model.MultiHealthReport{
		Repo:    repoName,
		Branch:  branch,
		Reports: reports,
	}

	for _, report := range reports {
		indexes = append(indexes, report.IndexFile)
		result.OpenTODOs += report.OpenTODOs
		result.CompletedTODOs += report.CompletedTODOs
		result.LintErrors += report.LintErrors
		result.LintWarnings += report.LintWarnings
		switch report.Status {
		case "error":
			result.IndexesWithErrors++
		case "warning":
			result.IndexesWithWarning++
		}
	}
	sort.Strings(indexes)
	result.IndexFiles = indexes
	result.Status = deriveMultiStatus(result)
	return result
}

func oldestOpenAges(ages []model.AgeRecord) []model.AgeRecord {
	var open []model.AgeRecord
	for _, age := range ages {
		if age.Todo.Status == model.StatusOpen {
			open = append(open, age)
		}
	}
	if len(open) > topOpenLimit {
		open = open[:topOpenLimit]
	}
	return open
}

func summarizeFindings(findings []model.LintFinding) []model.FindingCount {
	type key struct {
		severity string
		code     string
	}

	counts := map[key]int{}
	for _, finding := range findings {
		counts[key{severity: finding.Severity, code: finding.Code}]++
	}

	summary := make([]model.FindingCount, 0, len(counts))
	for k, count := range counts {
		summary = append(summary, model.FindingCount{
			Severity: k.severity,
			Code:     k.code,
			Count:    count,
		})
	}

	sort.Slice(summary, func(i, j int) bool {
		if severityRank(summary[i].Severity) == severityRank(summary[j].Severity) {
			if summary[i].Count == summary[j].Count {
				return summary[i].Code < summary[j].Code
			}
			return summary[i].Count > summary[j].Count
		}
		return severityRank(summary[i].Severity) < severityRank(summary[j].Severity)
	})
	return summary
}

func deriveStatus(errorsCount, warningsCount int, driftResult *model.DriftResult) string {
	switch {
	case errorsCount > 0:
		return "error"
	case warningsCount > 0:
		return "warning"
	case driftResult != nil && driftResult.TotalDifferenceRows > 0:
		return "warning"
	default:
		return "clean"
	}
}

func deriveMultiStatus(report model.MultiHealthReport) string {
	switch {
	case report.LintErrors > 0:
		return "error"
	case report.LintWarnings > 0 || report.IndexesWithWarning > 0:
		return "warning"
	default:
		return "clean"
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

func severityRank(severity string) int {
	switch severity {
	case "error":
		return 0
	case "warning":
		return 1
	default:
		return 2
	}
}
