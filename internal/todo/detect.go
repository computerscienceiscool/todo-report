package todo

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"todo-report/internal/model"
)

var styleRelatedCodes = map[string]bool{
	"broken_detail_link":   true,
	"invalid_checkbox":     true,
	"malformed_subtask_id": true,
	"malformed_todo_id":    true,
}

func DetectSnapshot(snapshot model.Snapshot) model.DetectReport {
	indexLayouts := detectIndexLayouts(snapshot)
	topLevelStyles := detectTopLevelStyles(snapshot.Items)
	subtaskStyles, features := detectSubtaskStyles(snapshot)
	styleFindings := summarizeStyleFindings(snapshot.Findings)

	compatibility := "compatible"
	if len(styleFindings) > 0 {
		compatibility = "unsupported"
	} else if len(indexLayouts) > 1 || len(topLevelStyles) > 1 || len(subtaskStyles) > 1 {
		compatibility = "compatible_with_warnings"
	}

	detailFileCount := 0
	subtaskCount := 0
	for _, item := range snapshot.Items {
		if item.DetailFile != "" {
			detailFileCount++
		}
		subtaskCount += len(snapshot.SubtasksByParent[item.TodoID])
	}

	return model.DetectReport{
		Repo:             snapshot.RepoName,
		Branch:           snapshot.Branch,
		IndexFile:        snapshot.IndexFile,
		TodoRoot:         snapshot.TodoRoot,
		Compatibility:    compatibility,
		IndexLayouts:     indexLayouts,
		TopLevelIDStyles: topLevelStyles,
		SubtaskIDStyles:  subtaskStyles,
		Features:         features,
		StyleFindings:    styleFindings,
		TopLevelCount:    len(snapshot.Items),
		DetailFileCount:  detailFileCount,
		SubtaskCount:     subtaskCount,
	}
}

func detectIndexLayouts(snapshot model.Snapshot) []string {
	content := snapshot.FileContents[snapshot.IndexFile]
	layouts := map[string]bool{}
	for _, line := range strings.Split(content, "\n") {
		if indexLineRE.MatchString(line) {
			layouts["checklist_lines"] = true
		}
		if indexTableRowRE.MatchString(line) {
			layouts["markdown_table_rows"] = true
		}
	}
	if len(layouts) == 0 {
		layouts["unknown"] = true
	}
	return sortedTrueKeys(layouts)
}

func detectTopLevelStyles(items []model.TodoItem) []string {
	styles := map[string]bool{}
	for _, item := range items {
		switch {
		case proquintRE.MatchString(item.TodoID):
			styles["proquint"] = true
		case legacyNumericRE.MatchString(item.TodoID):
			styles["numeric_legacy"] = true
		case legacyLetterRE.MatchString(item.TodoID):
			styles["letter_prefixed_legacy"] = true
		default:
			styles["other"] = true
		}
	}
	return sortedTrueKeys(styles)
}

func detectSubtaskStyles(snapshot model.Snapshot) ([]string, []string) {
	styles := map[string]bool{}
	features := map[string]bool{}

	for _, item := range snapshot.Items {
		parentStem := strings.TrimPrefix(item.TodoID, "TODO-")
		content := snapshot.FileContents[item.DetailFile]
		for _, line := range strings.Split(content, "\n") {
			if strings.Contains(line, "[~]") {
				features["approximate_checkboxes"] = true
			}
			for _, ref := range targetRefRE.FindAllStringSubmatch(line, -1) {
				if ref[1] != item.TodoID {
					features["cross_todo_references"] = true
				}
			}
			matches := subtaskLineRE.FindStringSubmatch(line)
			if matches == nil {
				continue
			}
			rawID := strings.Trim(strings.TrimSpace(matches[2]), "*`_")
			if strings.HasSuffix(rawID, ".") {
				features["trailing_dot_subtask_ids"] = true
			}
			styles[classifySubtaskStyle(parentStem, normalizeSubtaskID(rawID))] = true
		}
	}

	if len(styles) == 0 {
		styles["none"] = true
	}
	return sortedTrueKeys(styles), sortedTrueKeys(features)
}

func classifySubtaskStyle(parentStem, id string) string {
	switch {
	case regexp.MustCompile(`^\d+(?:\.\d+)*$`).MatchString(id):
		return "numeric_dotted"
	case parentStem != "" && strings.HasPrefix(id, parentStem+"."):
		return "parent_prefixed_dotted"
	case strings.Contains(id, "."):
		return "dotted_token"
	default:
		return "token"
	}
}

func summarizeStyleFindings(findings []model.LintFinding) []model.DetectFinding {
	type bucket struct {
		count    int
		examples []string
	}

	buckets := map[string]*bucket{}
	for _, finding := range findings {
		if !styleRelatedCodes[finding.Code] {
			continue
		}
		if buckets[finding.Code] == nil {
			buckets[finding.Code] = &bucket{}
		}
		buckets[finding.Code].count++
		if len(buckets[finding.Code].examples) < 3 {
			example := finding.File
			if finding.Line > 0 {
				example += ":" + strconv.Itoa(finding.Line)
			}
			buckets[finding.Code].examples = append(buckets[finding.Code].examples, example)
		}
	}

	var out []model.DetectFinding
	for code, bucket := range buckets {
		out = append(out, model.DetectFinding{
			Code:     code,
			Count:    bucket.count,
			Examples: bucket.examples,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out
}

func sortedTrueKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key, ok := range values {
		if ok {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}
