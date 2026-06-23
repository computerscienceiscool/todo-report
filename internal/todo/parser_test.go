package todo

import (
	"testing"

	"todo-report/internal/model"
)

func TestParseIndexSupportsProquintAndLegacyIDs(t *testing.T) {
	content := `# TODO Index

- [ ] TODO-binap - Lock outline (` + "`TODO/TODO-binap-readme-outline-lock.md`" + `)
001 - Add route support
S122 - Spelling check
`

	items, findings := ParseIndex("repo", "main", "abc123", "TODO/TODO.md", content)
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %#v", findings)
	}
	if items[0].TodoID != "TODO-binap" || items[1].TodoID != "001" || items[2].TodoID != "S122" {
		t.Fatalf("unexpected IDs: %#v", items)
	}
}

func TestParseDetailSupportsNestedSubtasks(t *testing.T) {
	content := `# TODO-binap

- [ ] binap.1 First
- [x] binap.2.1 Nested
- [ ] 2.1 Numeric legacy
`

	parent := sampleParent()
	subtasks, findings := ParseDetail("repo", "main", "abc123", parent, "TODO/TODO-binap.md", content)
	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %#v", findings)
	}
	if len(subtasks) != 3 {
		t.Fatalf("expected 3 subtasks, got %d", len(subtasks))
	}
	if subtasks[1].SubtaskID != "binap.2.1" || subtasks[2].SubtaskID != "2.1" {
		t.Fatalf("unexpected subtasks: %#v", subtasks)
	}
}

func sampleParent() model.TodoItem {
	return model.TodoItem{
		TodoID:     "TODO-binap",
		SourceFile: "TODO/TODO.md",
	}
}
