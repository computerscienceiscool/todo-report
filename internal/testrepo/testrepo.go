package testrepo

import (
	"os"
	"os/exec"
	"path/filepath"
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

	run(r.t, r.Dir, "git", "add", "TODO/TODO.md")
	run(r.t, r.Dir, "git", "add", "TODO")
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
