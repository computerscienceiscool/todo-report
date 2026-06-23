package age

import (
	"sort"
	"time"

	"todo-report/internal/gitrepo"
	"todo-report/internal/model"
	"todo-report/internal/todo"
)

func Compute(repo *gitrepo.Repo, snapshot model.Snapshot) ([]model.AgeRecord, error) {
	history, err := repo.ReverseLog(snapshot.Branch, "TODO/TODO.md")
	if err != nil {
		return nil, err
	}

	firstSeen := map[string]time.Time{}
	wanted := map[string]struct{}{}
	for _, item := range snapshot.Items {
		wanted[item.TodoID] = struct{}{}
	}

	for _, entry := range history {
		content, err := repo.ShowFile(entry.Commit, "TODO/TODO.md")
		if err != nil {
			continue
		}
		items, _ := todo.ParseIndex(snapshot.RepoName, snapshot.Branch, entry.Commit, "TODO/TODO.md", content)
		for _, item := range items {
			if _, ok := wanted[item.TodoID]; !ok {
				continue
			}
			if _, ok := firstSeen[item.TodoID]; ok {
				continue
			}
			firstSeen[item.TodoID] = entry.When
		}
	}

	now := time.Now()
	var records []model.AgeRecord
	for _, item := range snapshot.Items {
		when, ok := firstSeen[item.TodoID]
		if !ok {
			when = now
		}
		records = append(records, model.AgeRecord{
			Todo:      item,
			FirstSeen: when,
			AgeDays:   int(now.Sub(when).Hours() / 24),
		})
	}

	sort.Slice(records, func(i, j int) bool {
		if records[i].AgeDays == records[j].AgeDays {
			return records[i].Todo.TodoID < records[j].Todo.TodoID
		}
		return records[i].AgeDays > records[j].AgeDays
	})
	return records, nil
}
