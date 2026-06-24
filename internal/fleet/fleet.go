package fleet

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"todo-report/internal/model"
)

func LoadRepoList(listPath string) ([]string, error) {
	file, err := os.Open(listPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	baseDir := filepath.Dir(listPath)
	var repos []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}
		if line == "" {
			continue
		}
		resolved, err := resolveRepoPath(baseDir, line)
		if err != nil {
			return nil, err
		}
		repos = append(repos, resolved)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return repos, nil
}

func BuildHealthReport(branch, compareBranch, repoListFile string, entries []model.FleetHealthEntry) model.FleetHealthReport {
	report := model.FleetHealthReport{
		Branch:        branch,
		CompareBranch: compareBranch,
		RepoListFile:  repoListFile,
		RepoCount:     len(entries),
		Entries:       entries,
		Status:        "clean",
	}

	for _, entry := range entries {
		if entry.Error != "" {
			report.ErrorCount++
			report.Status = "error"
			continue
		}
		report.SuccessCount++
		report.OpenTODOs += entry.OpenTODOs
		report.CompletedTODOs += entry.CompletedTODOs
		report.LintErrors += entry.LintErrors
		report.LintWarnings += entry.LintWarnings
		report.DriftItems += entry.DriftItems
		switch entry.Status {
		case "error":
			report.Status = "error"
		case "warning":
			if report.Status != "error" {
				report.Status = "warning"
			}
		}
	}
	if report.Status == "clean" && (report.LintWarnings > 0 || report.DriftItems > 0) {
		report.Status = "warning"
	}
	sort.Slice(report.Entries, func(i, j int) bool {
		return report.Entries[i].RepoPath < report.Entries[j].RepoPath
	})
	return report
}

func FilterRepoPaths(repos, include, exclude []string) []string {
	if len(include) == 0 && len(exclude) == 0 {
		return repos
	}

	var filtered []string
	for _, repo := range repos {
		if !matchesAny(repo, include) && len(include) > 0 {
			continue
		}
		if matchesAny(repo, exclude) {
			continue
		}
		filtered = append(filtered, repo)
	}
	return filtered
}

func resolveRepoPath(baseDir, raw string) (string, error) {
	if strings.HasPrefix(raw, "~/") || raw == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if raw == "~" {
			raw = home
		} else {
			raw = filepath.Join(home, raw[2:])
		}
	}
	if !filepath.IsAbs(raw) {
		raw = filepath.Join(baseDir, raw)
	}
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", fmt.Errorf("resolve repo path %q: %w", raw, err)
	}
	return abs, nil
}

func matchesAny(value string, filters []string) bool {
	if len(filters) == 0 {
		return false
	}
	value = strings.ToLower(value)
	for _, filter := range filters {
		if filter == "" {
			continue
		}
		if strings.Contains(value, strings.ToLower(filter)) {
			return true
		}
	}
	return false
}
