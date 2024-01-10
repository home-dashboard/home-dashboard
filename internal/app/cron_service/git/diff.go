package git

import (
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
)

type DiffEntry struct {
	FileName string
	Action   merkletrie.Action
}

func Diff(commit *object.Commit, prevCommit *object.Commit) ([]DiffEntry, error) {
	currentTree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	prevTree, err := prevCommit.Tree()
	if err != nil {
		return nil, err
	}

	changes, err := currentTree.Diff(prevTree)
	if err != nil {
		return nil, err
	}

	diffs := make([]DiffEntry, 0)
	for _, change := range changes {
		// Ignore deleted files
		action, err := change.Action()
		if err != nil {
			return nil, err
		}

		fileName := ""
		if len(change.From.Name) > 0 {
			fileName = change.From.Name
		} else if len(change.To.Name) > 0 {
			fileName = change.To.Name
		} else {
			return nil, fmt.Errorf("malformed change: empty from and to")
		}

		diffs = append(diffs, DiffEntry{
			FileName: fileName,
			Action:   action,
		})
	}

	return diffs, nil
}
