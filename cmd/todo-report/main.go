package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"todo-report/internal/age"
	"todo-report/internal/drift"
	fleetcalc "todo-report/internal/fleet"
	"todo-report/internal/gitrepo"
	healthcalc "todo-report/internal/health"
	"todo-report/internal/lint"
	"todo-report/internal/model"
	"todo-report/internal/report"
	"todo-report/internal/todo"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return usageError()
	}

	switch args[0] {
	case "age":
		return runAge(args[1:])
	case "drift":
		return runDrift(args[1:])
	case "indexes":
		return runIndexes(args[1:])
	case "lint":
		return runLint(args[1:])
	case "fleet":
		return runFleet(args[1:])
	case "health":
		return runHealth(args[1:])
	default:
		return usageError()
	}
}

func runAge(args []string) error {
	fs := flag.NewFlagSet("age", flag.ContinueOnError)
	repoPath := fs.String("repo", ".", "path to git repo")
	branch := fs.String("branch", "", "branch to inspect")
	indexPath := fs.String("index", "TODO/TODO.md", "path to the authoritative TODO index, relative to repo root")
	format := fs.String("format", "text", "output format: text, markdown, json, tsv")
	jsonFlag := fs.Bool("json", false, "alias for --format json")
	if err := fs.Parse(args); err != nil {
		return err
	}
	formatValue, err := normalizeFormat(*format, *jsonFlag)
	if err != nil {
		return err
	}
	if *branch == "" {
		return errors.New("age requires --branch")
	}

	repo, err := gitrepo.Open(*repoPath)
	if err != nil {
		return err
	}

	snapshot, err := todo.LoadSnapshot(repo, *branch, *indexPath)
	if err != nil {
		return err
	}

	records, err := age.Compute(repo, snapshot)
	if err != nil {
		return err
	}

	out, err := report.RenderAge(records, formatValue)
	if err != nil {
		return err
	}
	fmt.Print(out)
	return nil
}

func runDrift(args []string) error {
	fs := flag.NewFlagSet("drift", flag.ContinueOnError)
	repoPath := fs.String("repo", ".", "path to git repo")
	branchA := fs.String("branch-a", "", "left branch")
	branchB := fs.String("branch-b", "", "right branch")
	indexPath := fs.String("index", "TODO/TODO.md", "path to the authoritative TODO index, relative to repo root")
	format := fs.String("format", "text", "output format: text, markdown, json, tsv")
	jsonFlag := fs.Bool("json", false, "alias for --format json")
	if err := fs.Parse(args); err != nil {
		return err
	}
	formatValue, err := normalizeFormat(*format, *jsonFlag)
	if err != nil {
		return err
	}
	if *branchA == "" || *branchB == "" {
		return errors.New("drift requires --branch-a and --branch-b")
	}

	repo, err := gitrepo.Open(*repoPath)
	if err != nil {
		return err
	}

	left, err := todo.LoadSnapshot(repo, *branchA, *indexPath)
	if err != nil {
		return err
	}
	right, err := todo.LoadSnapshot(repo, *branchB, *indexPath)
	if err != nil {
		return err
	}

	result := drift.Compare(left, right)
	out, err := report.RenderDrift(result, formatValue)
	if err != nil {
		return err
	}
	fmt.Print(out)
	return nil
}

func runLint(args []string) error {
	fs := flag.NewFlagSet("lint", flag.ContinueOnError)
	repoPath := fs.String("repo", ".", "path to git repo")
	branch := fs.String("branch", "", "branch to inspect")
	indexPath := fs.String("index", "TODO/TODO.md", "path to the authoritative TODO index, relative to repo root")
	format := fs.String("format", "text", "output format: text, markdown, json, tsv")
	jsonFlag := fs.Bool("json", false, "alias for --format json")
	if err := fs.Parse(args); err != nil {
		return err
	}
	formatValue, err := normalizeFormat(*format, *jsonFlag)
	if err != nil {
		return err
	}
	if *branch == "" {
		return errors.New("lint requires --branch")
	}

	repo, err := gitrepo.Open(*repoPath)
	if err != nil {
		return err
	}

	snapshot, err := todo.LoadSnapshot(repo, *branch, *indexPath)
	if err != nil {
		return err
	}

	findings := lint.Run(snapshot)
	out, err := report.RenderLint(snapshot, findings, formatValue)
	if err != nil {
		return err
	}
	fmt.Print(out)
	return nil
}

func runIndexes(args []string) error {
	fs := flag.NewFlagSet("indexes", flag.ContinueOnError)
	repoPath := fs.String("repo", ".", "path to git repo")
	branch := fs.String("branch", "", "branch to inspect")
	format := fs.String("format", "text", "output format: text, markdown, json, tsv")
	jsonFlag := fs.Bool("json", false, "alias for --format json")
	if err := fs.Parse(args); err != nil {
		return err
	}
	formatValue, err := normalizeFormat(*format, *jsonFlag)
	if err != nil {
		return err
	}
	if *branch == "" {
		return errors.New("indexes requires --branch")
	}

	repo, err := gitrepo.Open(*repoPath)
	if err != nil {
		return err
	}

	indexes, err := todo.DiscoverIndexes(repo, *branch)
	if err != nil {
		return err
	}

	out, err := report.RenderIndexes(indexes, formatValue)
	if err != nil {
		return err
	}
	fmt.Print(out)
	return nil
}

func runHealth(args []string) error {
	fs := flag.NewFlagSet("health", flag.ContinueOnError)
	repoPath := fs.String("repo", ".", "path to git repo")
	branch := fs.String("branch", "", "branch to inspect")
	indexPath := fs.String("index", "TODO/TODO.md", "path to the authoritative TODO index, relative to repo root")
	allIndexes := fs.Bool("all-indexes", false, "discover all TODO/TODO.md indexes and summarize them together")
	writeJSON := fs.String("write-json", "", "optional path to write the structured health report as JSON")
	format := fs.String("format", "text", "output format: text, markdown, json, tsv")
	jsonFlag := fs.Bool("json", false, "alias for --format json")
	compare := fs.String("compare", "", "optional branch to compare against")
	if err := fs.Parse(args); err != nil {
		return err
	}
	formatValue, err := normalizeFormat(*format, *jsonFlag)
	if err != nil {
		return err
	}
	if *branch == "" {
		return errors.New("health requires --branch")
	}

	repo, err := gitrepo.Open(*repoPath)
	if err != nil {
		return err
	}

	if *allIndexes {
		multiReport, err := loadMultiHealthReport(repo, *branch, *compare)
		if err != nil {
			return err
		}
		if *writeJSON != "" {
			if err := writeJSONFile(*writeJSON, multiReport); err != nil {
				return err
			}
		}
		out, err := report.RenderMultiHealth(multiReport, formatValue)
		if err != nil {
			return err
		}
		fmt.Print(out)
		return nil
	}

	reportData, err := loadHealthReport(repo, *branch, *indexPath, *compare)
	if err != nil {
		return err
	}
	if *writeJSON != "" {
		if err := writeJSONFile(*writeJSON, reportData); err != nil {
			return err
		}
	}

	out, err := report.RenderHealth(reportData, formatValue)
	if err != nil {
		return err
	}
	fmt.Print(out)
	return nil
}

func runFleet(args []string) error {
	if len(args) == 0 {
		return errors.New("fleet requires a subcommand")
	}
	switch args[0] {
	case "health":
		return runFleetHealth(args[1:])
	default:
		return fmt.Errorf("unsupported fleet subcommand %q", args[0])
	}
}

func runFleetHealth(args []string) error {
	fs := flag.NewFlagSet("fleet health", flag.ContinueOnError)
	repoList := fs.String("repo-list", "", "path to a newline-delimited repo list")
	branch := fs.String("branch", "", "branch to inspect")
	indexPath := fs.String("index", "TODO/TODO.md", "path to the authoritative TODO index, relative to repo root")
	allIndexes := fs.Bool("all-indexes", false, "discover all TODO/TODO.md indexes per repo and summarize them together")
	writeJSON := fs.String("write-json", "", "optional path to write the structured fleet report as JSON")
	format := fs.String("format", "text", "output format: text, markdown, json, tsv")
	jsonFlag := fs.Bool("json", false, "alias for --format json")
	compare := fs.String("compare", "", "optional branch to compare against")
	if err := fs.Parse(args); err != nil {
		return err
	}
	formatValue, err := normalizeFormat(*format, *jsonFlag)
	if err != nil {
		return err
	}
	if *repoList == "" {
		return errors.New("fleet health requires --repo-list")
	}
	if *branch == "" {
		return errors.New("fleet health requires --branch")
	}

	repos, err := fleetcalc.LoadRepoList(*repoList)
	if err != nil {
		return err
	}
	entries := make([]model.FleetHealthEntry, 0, len(repos))
	for _, repoPath := range repos {
		entry := model.FleetHealthEntry{
			RepoPath:   repoPath,
			IndexMode:  "single-index",
			IndexCount: 1,
		}
		if *allIndexes {
			entry.IndexMode = "all-indexes"
		}

		repo, err := gitrepo.Open(repoPath)
		if err != nil {
			entry.Repo = filepath.Base(repoPath)
			entry.Status = "error"
			entry.Error = err.Error()
			entries = append(entries, entry)
			continue
		}
		entry.Repo = repo.Name

		if *allIndexes {
			multiReport, err := loadMultiHealthReport(repo, *branch, *compare)
			if err != nil {
				entry.Status = "error"
				entry.Error = err.Error()
				entries = append(entries, entry)
				continue
			}
			entry.Status = multiReport.Status
			entry.IndexCount = len(multiReport.IndexFiles)
			entry.OpenTODOs = multiReport.OpenTODOs
			entry.CompletedTODOs = multiReport.CompletedTODOs
			entry.LintErrors = multiReport.LintErrors
			entry.LintWarnings = multiReport.LintWarnings
			entry.DriftItems = multiReport.DriftItems
			entry.IndexesWithErrors = multiReport.IndexesWithErrors
			entry.IndexesWithWarning = multiReport.IndexesWithWarning
			entry.MultiHealth = &multiReport
			entries = append(entries, entry)
			continue
		}

		healthReport, err := loadHealthReport(repo, *branch, *indexPath, *compare)
		if err != nil {
			entry.Status = "error"
			entry.Error = err.Error()
			entries = append(entries, entry)
			continue
		}
		entry.Status = healthReport.Status
		entry.OpenTODOs = healthReport.OpenTODOs
		entry.CompletedTODOs = healthReport.CompletedTODOs
		entry.LintErrors = healthReport.LintErrors
		entry.LintWarnings = healthReport.LintWarnings
		if healthReport.Drift != nil {
			entry.DriftItems = healthReport.Drift.TotalDifferenceRows
		}
		entry.IndexesWithErrors = boolToInt(healthReport.LintErrors > 0)
		entry.IndexesWithWarning = boolToInt(healthReport.LintWarnings > 0)
		entry.Health = &healthReport
		entries = append(entries, entry)
	}

	fleetReport := fleetcalc.BuildHealthReport(*branch, *compare, *repoList, entries)
	if *writeJSON != "" {
		if err := writeJSONFile(*writeJSON, fleetReport); err != nil {
			return err
		}
	}
	out, err := report.RenderFleetHealth(fleetReport, formatValue)
	if err != nil {
		return err
	}
	fmt.Print(out)
	return nil
}

func usageError() error {
	name := filepath.Base(os.Args[0])
	return fmt.Errorf("usage: %s <age|drift|indexes|lint|health|fleet> [flags]", name)
}

func normalizeFormat(format string, jsonFlag bool) (string, error) {
	if jsonFlag {
		format = "json"
	}
	format = strings.ToLower(strings.TrimSpace(format))
	switch format {
	case "text", "markdown", "json", "tsv":
		return format, nil
	default:
		return "", fmt.Errorf("unsupported format %q", format)
	}
}

func loadHealthReport(repo *gitrepo.Repo, branch, indexPath, compare string) (model.HealthReport, error) {
	snapshot, err := todo.LoadSnapshot(repo, branch, indexPath)
	if err != nil {
		return model.HealthReport{}, err
	}

	ages, err := age.Compute(repo, snapshot)
	if err != nil {
		return model.HealthReport{}, err
	}
	findings := lint.Run(snapshot)

	var driftResult *model.DriftResult
	if compare != "" {
		other, err := todo.LoadSnapshot(repo, compare, indexPath)
		if err != nil {
			return model.HealthReport{}, err
		}
		result := drift.Compare(snapshot, other)
		driftResult = &result
	}

	report := healthcalc.Build(snapshot, ages, findings, driftResult, compare)
	if compare != "" {
		report.PresentInCompare = true
	}
	return report, nil
}

func loadMultiHealthReport(repo *gitrepo.Repo, branch, compare string) (model.MultiHealthReport, error) {
	indexes, err := todo.DiscoverIndexes(repo, branch)
	if err != nil {
		return model.MultiHealthReport{}, err
	}
	var compareIndexes []string
	if compare != "" {
		compareIndexes, err = todo.DiscoverIndexes(repo, compare)
		if err != nil {
			return model.MultiHealthReport{}, err
		}
	}
	branchIndexSet := makeSet(indexes)
	compareIndexSet := makeSet(compareIndexes)
	onlyInBranch, onlyInCompare := diffSets(branchIndexSet, compareIndexSet)
	reports := make([]model.HealthReport, 0, len(indexes))
	for _, discovered := range indexes {
		compareForIndex := ""
		if compare != "" && compareIndexSet[discovered] {
			compareForIndex = compare
		}
		reportData, err := loadHealthReport(repo, branch, discovered, compareForIndex)
		if err != nil {
			return model.MultiHealthReport{}, err
		}
		reportData.PresentInCompare = compareForIndex == compare && compareForIndex != ""
		if compare != "" && !compareIndexSet[discovered] {
			reportData.Status = escalateStatus(reportData.Status)
		}
		reports = append(reports, reportData)
	}
	sort.Slice(reports, func(i, j int) bool { return reports[i].IndexFile < reports[j].IndexFile })
	return healthcalc.BuildMulti(repo.Name, branch, compare, reports, onlyInBranch, onlyInCompare), nil
}

func makeSet(values []string) map[string]bool {
	out := make(map[string]bool, len(values))
	for _, value := range values {
		out[value] = true
	}
	return out
}

func diffSets(left, right map[string]bool) (onlyLeft, onlyRight []string) {
	for value := range left {
		if !right[value] {
			onlyLeft = append(onlyLeft, value)
		}
	}
	for value := range right {
		if !left[value] {
			onlyRight = append(onlyRight, value)
		}
	}
	sort.Strings(onlyLeft)
	sort.Strings(onlyRight)
	return onlyLeft, onlyRight
}

func escalateStatus(status string) string {
	switch status {
	case "error", "warning":
		return status
	default:
		return "warning"
	}
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func writeJSONFile(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json for %s: %w", path, err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write json file %s: %w", path, err)
	}
	return nil
}
