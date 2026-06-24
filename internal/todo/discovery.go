package todo

import (
	"path"
	"sort"
	"strings"

	"todo-report/internal/gitrepo"
)

func DiscoverIndexes(repo *gitrepo.Repo, branch string) ([]string, error) {
	files, err := repo.ListAllFiles(branch)
	if err != nil {
		return nil, err
	}

	var indexes []string
	for _, file := range files {
		if file == "TODO/TODO.md" || strings.HasSuffix(file, "/TODO/TODO.md") {
			indexes = append(indexes, path.Clean(file))
		}
	}
	sort.Strings(indexes)
	return indexes, nil
}
