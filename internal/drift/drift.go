package drift

import (
	"sort"

	"todo-report/internal/model"
)

func Compare(left, right model.Snapshot) model.DriftResult {
	result := model.DriftResult{
		RepoName: left.RepoName,
		BranchA:  left.Branch,
		BranchB:  right.Branch,
	}

	for id, itemA := range left.ItemByID {
		itemB, ok := right.ItemByID[id]
		if !ok {
			result.OnlyInA = append(result.OnlyInA, id)
			continue
		}
		if itemA.Status == model.StatusCompleted && itemB.Status != model.StatusCompleted {
			result.CompletedOnlyInA = append(result.CompletedOnlyInA, id)
		}
		if itemA.Status != model.StatusCompleted && itemB.Status == model.StatusCompleted {
			result.CompletedOnlyInB = append(result.CompletedOnlyInB, id)
		}
		if itemA.DetailFile != "" && itemB.DetailFile == "" {
			result.DetailOnlyInA = append(result.DetailOnlyInA, id)
		}
		if itemA.DetailFile == "" && itemB.DetailFile != "" {
			result.DetailOnlyInB = append(result.DetailOnlyInB, id)
		}
		if itemA.Title != itemB.Title {
			result.OtherDifferences = append(result.OtherDifferences, model.DriftChange{
				Kind:    "title_changed",
				TodoID:  id,
				Details: itemA.Title + " <> " + itemB.Title,
			})
		}
	}
	for id := range right.ItemByID {
		if _, ok := left.ItemByID[id]; !ok {
			result.OnlyInB = append(result.OnlyInB, id)
		}
	}

	for target, subtaskA := range left.SubtaskByTarget {
		subtaskB, ok := right.SubtaskByTarget[target]
		if !ok {
			result.SubtaskOnlyInA = append(result.SubtaskOnlyInA, target)
			continue
		}
		if subtaskA.Status == model.StatusCompleted && subtaskB.Status != model.StatusCompleted {
			result.SubtaskCompletedA = append(result.SubtaskCompletedA, target)
		}
		if subtaskA.Status != model.StatusCompleted && subtaskB.Status == model.StatusCompleted {
			result.SubtaskCompletedB = append(result.SubtaskCompletedB, target)
		}
		if subtaskA.Title != subtaskB.Title {
			result.OtherDifferences = append(result.OtherDifferences, model.DriftChange{
				Kind:    "subtask_title_changed",
				Target:  target,
				Details: subtaskA.Title + " <> " + subtaskB.Title,
			})
		}
	}
	for target := range right.SubtaskByTarget {
		if _, ok := left.SubtaskByTarget[target]; !ok {
			result.SubtaskOnlyInB = append(result.SubtaskOnlyInB, target)
		}
	}

	sort.Strings(result.OnlyInA)
	sort.Strings(result.OnlyInB)
	sort.Strings(result.CompletedOnlyInA)
	sort.Strings(result.CompletedOnlyInB)
	sort.Strings(result.DetailOnlyInA)
	sort.Strings(result.DetailOnlyInB)
	sort.Strings(result.SubtaskOnlyInA)
	sort.Strings(result.SubtaskOnlyInB)
	sort.Strings(result.SubtaskCompletedA)
	sort.Strings(result.SubtaskCompletedB)
	sort.Slice(result.OtherDifferences, func(i, j int) bool {
		a := result.OtherDifferences[i]
		b := result.OtherDifferences[j]
		if a.Kind == b.Kind {
			if a.TodoID == b.TodoID {
				return a.Target < b.Target
			}
			return a.TodoID < b.TodoID
		}
		return a.Kind < b.Kind
	})

	result.TotalDifferenceRows = len(result.OnlyInA) + len(result.OnlyInB) + len(result.CompletedOnlyInA) +
		len(result.CompletedOnlyInB) + len(result.DetailOnlyInA) + len(result.DetailOnlyInB) +
		len(result.SubtaskOnlyInA) + len(result.SubtaskOnlyInB) + len(result.SubtaskCompletedA) +
		len(result.SubtaskCompletedB) + len(result.OtherDifferences)
	return result
}
