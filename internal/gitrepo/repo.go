package gitrepo

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Repo struct {
	Root string
	Name string
}

func Open(path string) (*Repo, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	root, err := runGit(abs, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Errorf("open repo: %w", err)
	}
	return &Repo{
		Root: strings.TrimSpace(root),
		Name: filepath.Base(strings.TrimSpace(root)),
	}, nil
}

func (r *Repo) BranchCommit(branch string) (string, error) {
	out, err := runGit(r.Root, "rev-parse", branch)
	if err != nil {
		return "", fmt.Errorf("resolve branch %q: %w", branch, err)
	}
	return strings.TrimSpace(out), nil
}

func (r *Repo) ShowFile(branch, path string) (string, error) {
	out, err := runGit(r.Root, "show", fmt.Sprintf("%s:%s", branch, path))
	if err != nil {
		return "", err
	}
	return out, nil
}

func (r *Repo) ListFiles(branch, prefix string) ([]string, error) {
	if strings.TrimSpace(prefix) == "" {
		return r.ListAllFiles(branch)
	}
	out, err := runGit(r.Root, "ls-tree", "-r", "--name-only", branch, "--", prefix)
	if err != nil {
		return nil, err
	}
	return splitFiles(out), nil
}

func (r *Repo) ListAllFiles(branch string) ([]string, error) {
	out, err := runGit(r.Root, "ls-tree", "-r", "--name-only", branch)
	if err != nil {
		return nil, err
	}
	return splitFiles(out), nil
}

func splitFiles(out string) []string {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	var files []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files
}

func (r *Repo) ReverseLog(branch, path string) ([]HistoryEntry, error) {
	out, err := runGit(r.Root, "log", "--reverse", "--format=%H\t%cI", branch, "--", path)
	if err != nil {
		return nil, err
	}
	var entries []HistoryEntry
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		when, err := time.Parse(time.RFC3339, parts[1])
		if err != nil {
			return nil, err
		}
		entries = append(entries, HistoryEntry{
			Commit: parts[0],
			When:   when,
		})
	}
	return entries, nil
}

type HistoryEntry struct {
	Commit string
	When   time.Time
}

func runGit(cwd string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = cwd
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return stdout.String(), nil
}
