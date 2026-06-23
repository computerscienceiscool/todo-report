package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"todo-report/internal/age"
	"todo-report/internal/drift"
	"todo-report/internal/gitrepo"
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
	case "lint":
		return runLint(args[1:])
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

	snapshot, err := todo.LoadSnapshot(repo, *branch)
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

	left, err := todo.LoadSnapshot(repo, *branchA)
	if err != nil {
		return err
	}
	right, err := todo.LoadSnapshot(repo, *branchB)
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

	snapshot, err := todo.LoadSnapshot(repo, *branch)
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

func runHealth(args []string) error {
	fs := flag.NewFlagSet("health", flag.ContinueOnError)
	repoPath := fs.String("repo", ".", "path to git repo")
	branch := fs.String("branch", "", "branch to inspect")
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

	snapshot, err := todo.LoadSnapshot(repo, *branch)
	if err != nil {
		return err
	}

	ages, err := age.Compute(repo, snapshot)
	if err != nil {
		return err
	}
	findings := lint.Run(snapshot)

	var driftResult *model.DriftResult
	if *compare != "" {
		other, err := todo.LoadSnapshot(repo, *compare)
		if err != nil {
			return err
		}
		result := drift.Compare(snapshot, other)
		driftResult = &result
	}

	out, err := report.RenderHealth(snapshot, ages, findings, driftResult, formatValue)
	if err != nil {
		return err
	}
	fmt.Print(out)
	return nil
}

func usageError() error {
	name := filepath.Base(os.Args[0])
	return fmt.Errorf("usage: %s <age|drift|lint|health> [flags]", name)
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
