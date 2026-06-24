package model

import "time"

type Status string

const (
	StatusOpen      Status = "open"
	StatusCompleted Status = "completed"
)

type TodoItem struct {
	Repo       string `json:"repo"`
	Branch     string `json:"branch"`
	CommitHash string `json:"commit_hash"`
	TodoID     string `json:"todo_id"`
	Title      string `json:"title"`
	Status     Status `json:"status"`
	SourceFile string `json:"source_file"`
	DetailFile string `json:"detail_file,omitempty"`
	Line       int    `json:"line"`
}

type Subtask struct {
	Repo       string `json:"repo"`
	Branch     string `json:"branch"`
	CommitHash string `json:"commit_hash"`
	ParentID   string `json:"parent_id"`
	SubtaskID  string `json:"subtask_id"`
	Title      string `json:"title"`
	Status     Status `json:"status"`
	SourceFile string `json:"source_file"`
	Line       int    `json:"line"`
}

func (s Subtask) Target() string {
	return s.ParentID + "/" + s.SubtaskID
}

type LintFinding struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	TodoID   string `json:"todo_id,omitempty"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Message  string `json:"message"`
}

type Snapshot struct {
	RepoName         string               `json:"repo_name"`
	RepoPath         string               `json:"repo_path"`
	Branch           string               `json:"branch"`
	CommitHash       string               `json:"commit_hash"`
	IndexFile        string               `json:"index_file"`
	TodoRoot         string               `json:"todo_root"`
	Items            []TodoItem           `json:"items"`
	ItemByID         map[string]TodoItem  `json:"-"`
	SubtasksByParent map[string][]Subtask `json:"subtasks_by_parent"`
	SubtaskByTarget  map[string]Subtask   `json:"-"`
	DetailFiles      map[string]string    `json:"detail_files"`
	OrphanDetail     []string             `json:"orphan_detail_files"`
	Findings         []LintFinding        `json:"findings"`
	Files            []string             `json:"files"`
	FileSet          map[string]struct{}  `json:"-"`
	FileContents     map[string]string    `json:"-"`
}

type AgeRecord struct {
	Todo      TodoItem  `json:"todo"`
	FirstSeen time.Time `json:"first_seen"`
	AgeDays   int       `json:"age_days"`
}

type DriftChange struct {
	Kind    string `json:"kind"`
	TodoID  string `json:"todo_id,omitempty"`
	Target  string `json:"target,omitempty"`
	Details string `json:"details"`
}

type DriftResult struct {
	RepoName            string        `json:"repo_name"`
	BranchA             string        `json:"branch_a"`
	BranchB             string        `json:"branch_b"`
	OnlyInA             []string      `json:"only_in_a"`
	OnlyInB             []string      `json:"only_in_b"`
	CompletedOnlyInA    []string      `json:"completed_only_in_a"`
	CompletedOnlyInB    []string      `json:"completed_only_in_b"`
	DetailOnlyInA       []string      `json:"detail_only_in_a"`
	DetailOnlyInB       []string      `json:"detail_only_in_b"`
	SubtaskOnlyInA      []string      `json:"subtask_only_in_a"`
	SubtaskOnlyInB      []string      `json:"subtask_only_in_b"`
	SubtaskCompletedA   []string      `json:"subtask_completed_only_in_a"`
	SubtaskCompletedB   []string      `json:"subtask_completed_only_in_b"`
	OtherDifferences    []DriftChange `json:"other_differences"`
	TotalDifferenceRows int           `json:"total_difference_rows"`
}

type FindingCount struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Count    int    `json:"count"`
}

type HealthReport struct {
	Repo              string         `json:"repo"`
	Branch            string         `json:"branch"`
	CompareBranch     string         `json:"compare_branch,omitempty"`
	IndexFile         string         `json:"index_file"`
	TodoRoot          string         `json:"todo_root"`
	PresentInBranch   bool           `json:"present_in_branch"`
	PresentInCompare  bool           `json:"present_in_compare,omitempty"`
	Status            string         `json:"status"`
	OpenTODOs         int            `json:"open_todos"`
	CompletedTODOs    int            `json:"completed_todos"`
	OldestOpen        *AgeRecord     `json:"oldest_open,omitempty"`
	OldestOpenItems   []AgeRecord    `json:"oldest_open_items,omitempty"`
	LintErrors        int            `json:"lint_errors"`
	LintWarnings      int            `json:"lint_warnings"`
	FindingSummary    []FindingCount `json:"finding_summary,omitempty"`
	Drift             *DriftResult   `json:"drift,omitempty"`
	Findings          []LintFinding  `json:"findings"`
	Age               []AgeRecord    `json:"age"`
	OrphanDetailFiles []string       `json:"orphan_detail_files"`
}

type MultiHealthReport struct {
	Repo                 string         `json:"repo"`
	Branch               string         `json:"branch"`
	CompareBranch        string         `json:"compare_branch,omitempty"`
	Status               string         `json:"status"`
	IndexFiles           []string       `json:"index_files"`
	Reports              []HealthReport `json:"reports"`
	IndexesOnlyInBranch  []string       `json:"indexes_only_in_branch,omitempty"`
	IndexesOnlyInCompare []string       `json:"indexes_only_in_compare,omitempty"`
	OpenTODOs            int            `json:"open_todos"`
	CompletedTODOs       int            `json:"completed_todos"`
	LintErrors           int            `json:"lint_errors"`
	LintWarnings         int            `json:"lint_warnings"`
	DriftItems           int            `json:"drift_items"`
	IndexesWithDrift     int            `json:"indexes_with_drift"`
	IndexesWithErrors    int            `json:"indexes_with_errors"`
	IndexesWithWarning   int            `json:"indexes_with_warnings"`
}
