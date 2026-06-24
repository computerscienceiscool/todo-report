package testrepo

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type Repo struct {
	Dir string
	t   *testing.T
}

func New(t *testing.T) *Repo {
	t.Helper()

	dir := t.TempDir()
	run(t, dir, "git", "init", "-b", "main")
	run(t, dir, "git", "config", "user.name", "Test User")
	run(t, dir, "git", "config", "user.email", "test@example.com")

	return &Repo{Dir: dir, t: t}
}

func (r *Repo) Write(path, content string) {
	r.t.Helper()

	full := filepath.Join(r.Dir, path)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		r.t.Fatalf("mkdir %s: %v", filepath.Dir(full), err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		r.t.Fatalf("write %s: %v", path, err)
	}
}

func (r *Repo) Commit(message, when string) {
	r.t.Helper()

	for _, path := range changedPaths(r.t, r.Dir) {
		run(r.t, r.Dir, "git", "add", path)
	}
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = r.Dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE="+when, "GIT_COMMITTER_DATE="+when)
	if out, err := cmd.CombinedOutput(); err != nil {
		r.t.Fatalf("git commit: %v\n%s", err, string(out))
	}
}

func (r *Repo) CheckoutNew(branch string) {
	r.t.Helper()
	run(r.t, r.Dir, "git", "checkout", "-b", branch)
}

func (r *Repo) Checkout(branch string) {
	r.t.Helper()
	run(r.t, r.Dir, "git", "checkout", branch)
}

func run(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, string(out))
	}
}

func changedPaths(t *testing.T, dir string) []string {
	t.Helper()

	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("git status --porcelain: %v\n%s", err, stderr.String())
	}

	var paths []string
	for _, line := range strings.Split(stdout.String(), "\n") {
		if strings.TrimSpace(line) == "" || len(line) < 4 {
			continue
		}
		paths = append(paths, strings.TrimSpace(line[3:]))
	}
	return paths
}
