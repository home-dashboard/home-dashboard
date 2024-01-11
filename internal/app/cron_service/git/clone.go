package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"os"
)

func CloneAndCheckout(url string, path string, branch string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	} else if err := os.RemoveAll(path); err != nil {
		return err
	}

	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:           url,
		ReferenceName: plumbing.ReferenceName("refs/heads/" + branch),
		SingleBranch:  true,
		Depth:         1,
	})
	if err != nil {
		return err
	}

	return nil
}
