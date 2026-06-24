package todo

import (
	"testing"

	"todo-report/internal/model"
)

func TestParseIndexSupportsProquintAndLegacyIDs(t *testing.T) {
	content := `# TODO Index

- [ ] TODO-binap - Lock outline (` + "`TODO/TODO-binap-readme-outline-lock.md`" + `)
barez - Bare proquint
001 - Add route support
S122 - Spelling check
`

	items, findings := ParseIndex("repo", "main", "abc123", "TODO/TODO.md", content)
	if len(items) != 4 {
		t.Fatalf("expected 4 items, got %d", len(items))
	}
	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %#v", findings)
	}
	if items[0].TodoID != "TODO-binap" || items[1].TodoID != "barez" || items[2].TodoID != "001" || items[3].TodoID != "S122" {
		t.Fatalf("unexpected IDs: %#v", items)
	}
}

func TestParseIndexSupportsFilenameStemIDs(t *testing.T) {
	content := `# Storm TODO

- [ ] 026-planning-group-workspace-mvp.md Planning group workspace tool MVP (shared docs, decisions, collaboration)
- [x] 014-change-review-gate.md Change review gate (diff/approve/apply/commit) for file edits
`

	items, findings := ParseIndex("repo", "main", "abc123", "TODO/TODO.md", content)
	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %#v", findings)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].TodoID != "026-planning-group-workspace-mvp.md" || items[0].DetailFile != "TODO/026-planning-group-workspace-mvp.md" {
		t.Fatalf("unexpected first item %#v", items[0])
	}
	if items[1].TodoID != "014-change-review-gate.md" || items[1].Status != model.StatusCompleted {
		t.Fatalf("unexpected second item %#v", items[1])
	}
}

func TestParseIndexSupportsMarkdownTableRowsAndRelativeLinks(t *testing.T) {
	content := `# TODO queue

| Handle | Mint date | Title | Prior alias |
|---|---|---|---|
| [TODO-hipak](./TODO-hipak-local.md) | 2026-05-25 | Local title **(implemented)** | — |
| [TODO-bisur](../../../simulations/SIM-rakot-group-session/protocols/group-session.d/TODO/TODO-bisur-group-transport-envelope.md) | 2026-05-01 | Cross-tree title | — |
`

	items, findings := ParseIndex("repo", "main", "abc123", "protocols/wire-lab.d/TODO/TODO.md", content)
	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %#v", findings)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].DetailFile != "protocols/wire-lab.d/TODO/TODO-hipak-local.md" {
		t.Fatalf("unexpected resolved local detail path %q", items[0].DetailFile)
	}
	if items[1].DetailFile != "simulations/SIM-rakot-group-session/protocols/group-session.d/TODO/TODO-bisur-group-transport-envelope.md" {
		t.Fatalf("unexpected resolved cross-tree detail path %q", items[1].DetailFile)
	}
	if items[0].Title != "Local title (implemented)" {
		t.Fatalf("unexpected cleaned title %q", items[0].Title)
	}
}

func TestParseDetailSupportsNestedSubtasks(t *testing.T) {
	content := `# TODO-binap

- [ ] binap.1 First
- [x] binap.2.1 Nested
- [ ] 2.1 Numeric legacy
- [ ] UT-155.a Wire-lab style
- [ ] UT-PSTK-origin Single token
- [~] 1. Approximate
`

	parent := sampleParent()
	subtasks, findings := ParseDetail("repo", "main", "abc123", parent, "TODO/TODO-binap.md", content)
	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %#v", findings)
	}
	if len(subtasks) != 6 {
		t.Fatalf("expected 6 subtasks, got %d", len(subtasks))
	}
	if subtasks[1].SubtaskID != "binap.2.1" || subtasks[2].SubtaskID != "2.1" || subtasks[3].SubtaskID != "UT-155.a" || subtasks[4].SubtaskID != "UT-PSTK-origin" || subtasks[5].SubtaskID != "1" {
		t.Fatalf("unexpected subtasks: %#v", subtasks)
	}
	if subtasks[5].Status != model.StatusOpen {
		t.Fatalf("expected [~] subtask to remain open, got %q", subtasks[5].Status)
	}
}

func TestParseIndexReportsBrokenLinksAndBadCheckboxes(t *testing.T) {
	content := `# TODO Index

- [q] TODO-binap - Lock outline (TODO/TODO-binap.md)
TODO-badi - Invalid top-level (` + "`TODO/TODO-badi.md`" + `)
`

	_, findings := ParseIndex("repo", "main", "abc123", "TODO/TODO.md", content)
	codes := findCodes(findings)
	for _, want := range []string{"broken_detail_link", "invalid_checkbox", "malformed_todo_id"} {
		if !codes[want] {
			t.Fatalf("expected code %q in findings %#v", want, findings)
		}
	}
}

func TestParseDetailReportsMissingReferencedSubtask(t *testing.T) {
	content := `# TODO-binap

- [ ] binap.1 First
Depends on TODO-binap/binap.9
Blocked by TODO-other/TE-43
`

	_, findings := ParseDetail("repo", "main", "abc123", sampleParent(), "TODO/TODO-binap.md", content)
	codes := findCodes(findings)
	if !codes["referenced_subtask_not_found"] {
		t.Fatalf("expected referenced_subtask_not_found in %#v", findings)
	}
}

func TestParseDetailIgnoresCrossTodoReferences(t *testing.T) {
	content := `# TODO-binap

- [ ] binap.1 First
Blocked by TODO-other/TE-43
`

	_, findings := ParseDetail("repo", "main", "abc123", sampleParent(), "TODO/TODO-binap.md", content)
	codes := findCodes(findings)
	if codes["referenced_subtask_not_found"] {
		t.Fatalf("expected cross-TODO references to be ignored, got %#v", findings)
	}
}

func TestParseDetailReportsMalformedSubtaskID(t *testing.T) {
	content := `# TODO-binap

- [ ] (binap.one) Bad subtask id
- [ ] valid.1 Good subtask id
`

	subtasks, findings := ParseDetail("repo", "main", "abc123", sampleParent(), "TODO/TODO-binap.md", content)
	codes := findCodes(findings)
	if !codes["malformed_subtask_id"] {
		t.Fatalf("expected malformed_subtask_id in %#v", findings)
	}
	if len(subtasks) != 1 || subtasks[0].SubtaskID != "valid.1" {
		t.Fatalf("expected only valid subtask to parse, got %#v", subtasks)
	}
}

func sampleParent() model.TodoItem {
	return model.TodoItem{
		TodoID:     "TODO-binap",
		SourceFile: "TODO/TODO.md",
	}
}

func findCodes(findings []model.LintFinding) map[string]bool {
	out := make(map[string]bool, len(findings))
	for _, finding := range findings {
		out[finding.Code] = true
	}
	return out
}
