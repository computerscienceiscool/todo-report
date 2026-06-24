package todo

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"

	"todo-report/internal/gitrepo"
	"todo-report/internal/model"
)

var (
	indexLineRE     = regexp.MustCompile(`^\s*(?:-\s*)?(?:\[( |x|X|~)\]\s+)?([A-Za-z0-9-]+)\s+-\s+(.+?)(?:\s+\(` + "`([^`]+)`" + `\))?\s*$`)
	indexFileStemRE = regexp.MustCompile(`^\s*(?:-\s*)?(?:\[( |x|X|~)\]\s+)?([A-Za-z0-9][A-Za-z0-9.-]*\.md)\s+(.+?)\s*$`)
	subtaskLineRE   = regexp.MustCompile(`^\s*-\s+\[( |x|X|~)\]\s+(\S+)(?:\s+(.+?))?\s*$`)
	indexTableRowRE = regexp.MustCompile(`^\|\s*\[([A-Za-z0-9-]+)\]\(([^)]+)\)\s*\|\s*([^|]*)\|\s*(.*?)\s*\|\s*([^|]*)\|\s*$`)
	badCheckboxRE   = regexp.MustCompile(`^\s*-\s+\[[^ xX~]\]`)
	targetRefRE     = regexp.MustCompile(`\b(TODO-[a-z]{5})/([A-Za-z0-9-]+(?:\.[A-Za-z0-9-]+)*)\b`)
	proquintRE      = regexp.MustCompile(`^TODO-[a-z]{5}$`)
	bareProquintRE  = regexp.MustCompile(`^[a-z]{5}$`)
	legacyNumericRE = regexp.MustCompile(`^\d{3,4}$`)
	legacyLetterRE  = regexp.MustCompile(`^[A-Za-z]\d{2,4}$`)
	filenameStemRE  = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.-]*\.md$`)
)

func LoadSnapshot(repo *gitrepo.Repo, branch, indexPath string) (model.Snapshot, error) {
	indexPath = normalizeIndexPath(indexPath)
	todoRoot := path.Dir(indexPath)

	commit, err := repo.BranchCommit(branch)
	if err != nil {
		return model.Snapshot{}, err
	}

	indexContent, err := repo.ShowFile(branch, indexPath)
	if err != nil {
		return model.Snapshot{}, fmt.Errorf("load %s from %s: %w", indexPath, branch, err)
	}

	items, findings := ParseIndex(repo.Name, branch, commit, indexPath, indexContent)
	files, err := repo.ListFiles(branch, todoRoot)
	if err != nil {
		return model.Snapshot{}, err
	}

	snapshot := model.Snapshot{
		RepoName:         repo.Name,
		RepoPath:         repo.Root,
		Branch:           branch,
		CommitHash:       commit,
		IndexFile:        indexPath,
		TodoRoot:         todoRoot,
		Items:            items,
		ItemByID:         make(map[string]model.TodoItem, len(items)),
		IndexStatusByID:  make(map[string]model.Status, len(items)),
		SubtasksByParent: make(map[string][]model.Subtask),
		SubtaskByTarget:  make(map[string]model.Subtask),
		DetailFiles:      make(map[string]string),
		Findings:         findings,
		Files:            files,
		FileSet:          make(map[string]struct{}, len(files)),
		FileContents:     map[string]string{indexPath: indexContent},
	}

	for _, file := range files {
		snapshot.FileSet[file] = struct{}{}
	}
	for _, item := range items {
		snapshot.ItemByID[item.TodoID] = item
		snapshot.IndexStatusByID[item.TodoID] = item.Status
		if item.DetailFile != "" {
			snapshot.DetailFiles[item.TodoID] = item.DetailFile
		}
	}

	referenced := map[string]struct{}{}
	for i, item := range items {
		if item.DetailFile == "" {
			continue
		}
		referenced[item.DetailFile] = struct{}{}
		content, err := repo.ShowFile(branch, item.DetailFile)
		if err != nil {
			snapshot.Findings = append(snapshot.Findings, model.LintFinding{
				Severity: "error",
				Code:     "missing_detail_file",
				TodoID:   item.TodoID,
				File:     item.SourceFile,
				Line:     item.Line,
				Message:  fmt.Sprintf("Detail file %q does not exist on branch %s.", item.DetailFile, branch),
			})
			continue
		}
		snapshot.FileContents[item.DetailFile] = content
		item.Status = deriveTodoStatus(item.Status, content)
		snapshot.Items[i] = item
		snapshot.ItemByID[item.TodoID] = item
		snapshot.DetailFiles[item.TodoID] = item.DetailFile
		subtasks, detailFindings := ParseDetail(repo.Name, branch, commit, item, item.DetailFile, content)
		snapshot.SubtasksByParent[item.TodoID] = subtasks
		for _, subtask := range subtasks {
			snapshot.SubtaskByTarget[subtask.Target()] = subtask
		}
		snapshot.Findings = append(snapshot.Findings, detailFindings...)
	}

	for _, file := range files {
		if file == indexPath {
			continue
		}
		if !strings.HasPrefix(file, todoRoot+"/TODO-") || !strings.HasSuffix(file, ".md") {
			continue
		}
		if _, ok := referenced[file]; ok {
			continue
		}
		snapshot.OrphanDetail = append(snapshot.OrphanDetail, file)
	}
	sort.Strings(snapshot.OrphanDetail)
	return snapshot, nil
}

func ParseIndex(repoName, branch, commit, file, content string) ([]model.TodoItem, []model.LintFinding) {
	var items []model.TodoItem
	var findings []model.LintFinding
	seen := map[string]int{}
	indexDir := path.Dir(file)

	for i, line := range strings.Split(content, "\n") {
		lineNo := i + 1
		if !strings.HasPrefix(strings.TrimSpace(line), "|") && strings.Contains(line, "TODO/") && !strings.Contains(line, "(`TODO/") && !strings.Contains(line, "(`") {
			if strings.Contains(line, "TODO/TODO-") && !strings.Contains(line, "(`TODO/") {
				findings = append(findings, model.LintFinding{
					Severity: "error",
					Code:     "broken_detail_link",
					File:     file,
					Line:     lineNo,
					Message:  "Detail file link must use the (`path`) backtick form.",
				})
			}
		}
		if badCheckboxRE.MatchString(line) {
			findings = append(findings, model.LintFinding{
				Severity: "error",
				Code:     "invalid_checkbox",
				File:     file,
				Line:     lineNo,
				Message:  "Checkbox must be [ ] or [x].",
			})
		}
		matches := indexLineRE.FindStringSubmatch(line)
		if matches != nil {
			status := parseCheckboxStatus(matches[1])
			todoID := matches[2]
			title := strings.TrimSpace(matches[3])
			detailFile := resolveDetailPath(indexDir, strings.TrimSpace(matches[4]))

			findings = append(findings, validateTopLevelItem(todoID, detailFile, file, lineNo, seen)...)
			if _, ok := seen[todoID]; !ok {
				seen[todoID] = lineNo
			}

			items = append(items, model.TodoItem{
				Repo:       repoName,
				Branch:     branch,
				CommitHash: commit,
				TodoID:     todoID,
				Title:      title,
				Status:     status,
				SourceFile: file,
				DetailFile: detailFile,
				Line:       lineNo,
			})
			continue
		}

		stemMatches := indexFileStemRE.FindStringSubmatch(line)
		if stemMatches != nil {
			status := parseCheckboxStatus(stemMatches[1])
			todoID := stemMatches[2]
			title := strings.TrimSpace(stemMatches[3])
			detailFile := resolveDetailPath(indexDir, todoID)

			findings = append(findings, validateTopLevelItem(todoID, detailFile, file, lineNo, seen)...)
			if _, ok := seen[todoID]; !ok {
				seen[todoID] = lineNo
			}

			items = append(items, model.TodoItem{
				Repo:       repoName,
				Branch:     branch,
				CommitHash: commit,
				TodoID:     todoID,
				Title:      title,
				Status:     status,
				SourceFile: file,
				DetailFile: detailFile,
				Line:       lineNo,
			})
			continue
		}

		rowMatches := indexTableRowRE.FindStringSubmatch(line)
		if rowMatches == nil {
			continue
		}
		todoID := rowMatches[1]
		detailFile := resolveDetailPath(indexDir, strings.TrimSpace(rowMatches[2]))
		title := cleanTableTitle(strings.TrimSpace(rowMatches[4]))

		findings = append(findings, validateTopLevelItem(todoID, detailFile, file, lineNo, seen)...)
		if _, ok := seen[todoID]; !ok {
			seen[todoID] = lineNo
		}

		items = append(items, model.TodoItem{
			Repo:       repoName,
			Branch:     branch,
			CommitHash: commit,
			TodoID:     todoID,
			Title:      title,
			Status:     model.StatusOpen,
			SourceFile: file,
			DetailFile: detailFile,
			Line:       lineNo,
		})
	}

	return items, findings
}

func ParseDetail(repoName, branch, commit string, parent model.TodoItem, file, content string) ([]model.Subtask, []model.LintFinding) {
	var subtasks []model.Subtask
	var findings []model.LintFinding
	seen := map[string]int{}
	refs := map[string]int{}

	for i, line := range strings.Split(content, "\n") {
		lineNo := i + 1
		if badCheckboxRE.MatchString(line) {
			findings = append(findings, model.LintFinding{
				Severity: "error",
				Code:     "invalid_checkbox",
				TodoID:   parent.TodoID,
				File:     file,
				Line:     lineNo,
				Message:  "Checkbox must be [ ] or [x].",
			})
		}
		for _, ref := range targetRefRE.FindAllStringSubmatch(line, -1) {
			if ref[1] != parent.TodoID {
				continue
			}
			target := ref[1] + "/" + normalizeSubtaskID(ref[2])
			refs[target] = lineNo
		}
		matches := subtaskLineRE.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		status := model.StatusOpen
		if strings.EqualFold(matches[1], "x") {
			status = model.StatusCompleted
		}
		subtaskID := normalizeSubtaskID(matches[2])
		title := strings.TrimSpace(matches[3])
		if !validSubtaskID(subtaskID) {
			findings = append(findings, model.LintFinding{
				Severity: "error",
				Code:     "malformed_subtask_id",
				TodoID:   parent.TodoID,
				File:     file,
				Line:     lineNo,
				Message:  fmt.Sprintf("Subtask ID %q is malformed.", subtaskID),
			})
			continue
		}
		if firstLine, ok := seen[subtaskID]; ok {
			findings = append(findings, model.LintFinding{
				Severity: "error",
				Code:     "duplicate_subtask_id",
				TodoID:   parent.TodoID,
				File:     file,
				Line:     lineNo,
				Message:  fmt.Sprintf("Subtask ID %q was already declared on line %d.", subtaskID, firstLine),
			})
		} else {
			seen[subtaskID] = lineNo
		}

		subtasks = append(subtasks, model.Subtask{
			Repo:       repoName,
			Branch:     branch,
			CommitHash: commit,
			ParentID:   parent.TodoID,
			SubtaskID:  subtaskID,
			Title:      title,
			Status:     status,
			SourceFile: file,
			Line:       lineNo,
		})
	}

	known := map[string]struct{}{}
	for _, subtask := range subtasks {
		known[subtask.Target()] = struct{}{}
	}
	for target, lineNo := range refs {
		if _, ok := known[target]; ok {
			continue
		}
		findings = append(findings, model.LintFinding{
			Severity: "error",
			Code:     "referenced_subtask_not_found",
			TodoID:   parent.TodoID,
			File:     file,
			Line:     lineNo,
			Message:  fmt.Sprintf("Referenced subtask %q was not found in the loaded detail files.", target),
		})
	}

	return subtasks, findings
}

func validTopLevelID(id string) bool {
	return proquintRE.MatchString(id) || bareProquintRE.MatchString(id) || legacyNumericRE.MatchString(id) || legacyLetterRE.MatchString(id) || filenameStemRE.MatchString(id)
}

func validSubtaskID(id string) bool {
	id = normalizeSubtaskID(id)
	if id == "" {
		return false
	}
	if regexp.MustCompile(`^\d+(?:\.\d+)*$`).MatchString(id) {
		return true
	}
	return regexp.MustCompile(`^[A-Za-z0-9-]+(?:\.[A-Za-z0-9-]+)*$`).MatchString(id)
}

func normalizeIndexPath(indexPath string) string {
	indexPath = strings.TrimSpace(indexPath)
	if indexPath == "" {
		return "TODO/TODO.md"
	}
	return path.Clean(indexPath)
}

func resolveDetailPath(indexDir, link string) string {
	link = strings.TrimSpace(link)
	if link == "" {
		return ""
	}
	if strings.HasPrefix(link, "./") || strings.HasPrefix(link, "../") || !strings.Contains(link, "/") {
		return path.Clean(path.Join(indexDir, link))
	}
	return path.Clean(link)
}

func validateTopLevelItem(todoID, detailFile, file string, lineNo int, seen map[string]int) []model.LintFinding {
	var findings []model.LintFinding
	if !validTopLevelID(todoID) {
		findings = append(findings, model.LintFinding{
			Severity: "error",
			Code:     "malformed_todo_id",
			TodoID:   todoID,
			File:     file,
			Line:     lineNo,
			Message:  fmt.Sprintf("TODO ID %q is not a supported top-level TODO ID style.", todoID),
		})
	}
	if firstLine, ok := seen[todoID]; ok {
		findings = append(findings, model.LintFinding{
			Severity: "error",
			Code:     "duplicate_id",
			TodoID:   todoID,
			File:     file,
			Line:     lineNo,
			Message:  fmt.Sprintf("TODO ID %q was already declared on line %d.", todoID, firstLine),
		})
	}
	if detailFile != "" && path.Ext(detailFile) != ".md" {
		findings = append(findings, model.LintFinding{
			Severity: "error",
			Code:     "broken_detail_link",
			TodoID:   todoID,
			File:     file,
			Line:     lineNo,
			Message:  fmt.Sprintf("Detail file %q must end with .md.", detailFile),
		})
	}
	return findings
}

func parseCheckboxStatus(raw string) model.Status {
	if strings.EqualFold(raw, "x") {
		return model.StatusCompleted
	}
	return model.StatusOpen
}

func cleanTableTitle(title string) string {
	title = strings.TrimSpace(title)
	title = strings.ReplaceAll(title, "**", "")
	return title
}

func deriveTodoStatus(current model.Status, content string) model.Status {
	if current == model.StatusCompleted {
		return current
	}
	statusText := extractStatusText(content)
	switch {
	case strings.HasPrefix(statusText, "implemented"),
		strings.HasPrefix(statusText, "closed"),
		strings.HasPrefix(statusText, "retired"),
		strings.HasPrefix(statusText, "deferred"),
		strings.HasPrefix(statusText, "folded"):
		return model.StatusCompleted
	default:
		return model.StatusOpen
	}
}

func normalizeSubtaskID(id string) string {
	id = strings.Trim(strings.TrimSpace(id), "*`_")
	return strings.TrimRight(id, ".")
}

func extractStatusText(content string) string {
	lines := strings.Split(content, "\n")
	inStatus := false
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "## Status" {
			inStatus = true
			continue
		}
		if !inStatus {
			continue
		}
		if strings.HasPrefix(line, "## ") {
			break
		}
		if line == "" {
			continue
		}
		return strings.ToLower(line)
	}
	return ""
}
