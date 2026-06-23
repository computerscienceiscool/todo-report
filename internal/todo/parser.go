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
	indexLineRE     = regexp.MustCompile(`^\s*(?:-\s*)?(?:\[( |x|X)\]\s+)?([A-Za-z0-9-]+)\s+-\s+(.+?)(?:\s+\(` + "`([^`]+)`" + `\))?\s*$`)
	subtaskLineRE   = regexp.MustCompile(`^\s*-\s+\[( |x|X)\]\s+([A-Za-z0-9]+(?:\.\d+)+)\s+(.+?)\s*$`)
	badCheckboxRE   = regexp.MustCompile(`^\s*-\s+\[[^ xX]\]`)
	targetRefRE     = regexp.MustCompile(`\b(TODO-[a-z]{5})/([A-Za-z0-9]+(?:\.\d+)+)\b`)
	proquintRE      = regexp.MustCompile(`^TODO-[a-z]{5}$`)
	legacyNumericRE = regexp.MustCompile(`^\d{3,4}$`)
	legacyLetterRE  = regexp.MustCompile(`^[A-Za-z]\d{2,4}$`)
)

func LoadSnapshot(repo *gitrepo.Repo, branch string) (model.Snapshot, error) {
	commit, err := repo.BranchCommit(branch)
	if err != nil {
		return model.Snapshot{}, err
	}

	indexContent, err := repo.ShowFile(branch, "TODO/TODO.md")
	if err != nil {
		return model.Snapshot{}, fmt.Errorf("load TODO/TODO.md from %s: %w", branch, err)
	}

	items, findings := ParseIndex(repo.Name, branch, commit, "TODO/TODO.md", indexContent)
	files, err := repo.ListFiles(branch, "TODO")
	if err != nil {
		return model.Snapshot{}, err
	}

	snapshot := model.Snapshot{
		RepoName:         repo.Name,
		RepoPath:         repo.Root,
		Branch:           branch,
		CommitHash:       commit,
		Items:            items,
		ItemByID:         make(map[string]model.TodoItem, len(items)),
		SubtasksByParent: make(map[string][]model.Subtask),
		SubtaskByTarget:  make(map[string]model.Subtask),
		DetailFiles:      make(map[string]string),
		Findings:         findings,
		Files:            files,
		FileSet:          make(map[string]struct{}, len(files)),
		FileContents:     map[string]string{"TODO/TODO.md": indexContent},
	}

	for _, file := range files {
		snapshot.FileSet[file] = struct{}{}
	}
	for _, item := range items {
		snapshot.ItemByID[item.TodoID] = item
		if item.DetailFile != "" {
			snapshot.DetailFiles[item.TodoID] = item.DetailFile
		}
	}

	referenced := map[string]struct{}{}
	for _, item := range items {
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
		subtasks, detailFindings := ParseDetail(repo.Name, branch, commit, item, item.DetailFile, content)
		snapshot.SubtasksByParent[item.TodoID] = subtasks
		for _, subtask := range subtasks {
			snapshot.SubtaskByTarget[subtask.Target()] = subtask
		}
		snapshot.Findings = append(snapshot.Findings, detailFindings...)
	}

	for _, file := range files {
		if !strings.HasPrefix(file, "TODO/TODO-") || !strings.HasSuffix(file, ".md") {
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

	for i, line := range strings.Split(content, "\n") {
		lineNo := i + 1
		if strings.Contains(line, "TODO/") && !strings.Contains(line, "(`TODO/") && !strings.Contains(line, "(`") {
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
		if matches == nil {
			continue
		}

		status := model.StatusOpen
		if strings.EqualFold(matches[1], "x") {
			status = model.StatusCompleted
		}
		todoID := matches[2]
		title := strings.TrimSpace(matches[3])
		detailFile := strings.TrimSpace(matches[4])

		if !validTopLevelID(todoID) {
			findings = append(findings, model.LintFinding{
				Severity: "error",
				Code:     "malformed_todo_id",
				TodoID:   todoID,
				File:     file,
				Line:     lineNo,
				Message:  fmt.Sprintf("TODO ID %q is not a supported proquint or legacy ID.", todoID),
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
		} else {
			seen[todoID] = lineNo
		}
		if detailFile != "" && (!strings.HasPrefix(detailFile, "TODO/") || path.Ext(detailFile) != ".md") {
			findings = append(findings, model.LintFinding{
				Severity: "error",
				Code:     "broken_detail_link",
				TodoID:   todoID,
				File:     file,
				Line:     lineNo,
				Message:  fmt.Sprintf("Detail file %q must live under TODO/ and end with .md.", detailFile),
			})
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
			refs[ref[1]+"/"+ref[2]] = lineNo
		}
		matches := subtaskLineRE.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		status := model.StatusOpen
		if strings.EqualFold(matches[1], "x") {
			status = model.StatusCompleted
		}
		subtaskID := matches[2]
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
	return proquintRE.MatchString(id) || legacyNumericRE.MatchString(id) || legacyLetterRE.MatchString(id)
}

func validSubtaskID(id string) bool {
	if regexp.MustCompile(`^\d+(?:\.\d+)*$`).MatchString(id) {
		return true
	}
	return regexp.MustCompile(`^[A-Za-z0-9]+(?:\.\d+)+$`).MatchString(id)
}
