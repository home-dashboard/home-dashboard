package project

import (
	"encoding/json"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/samber/lo"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/constants"
	git2 "github.com/siaikin/home-dashboard/internal/app/cron_service/git"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/model"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/project/db_utils"
)

func Exec(project model.Project, branchName string) error {
	repo, err := git.PlainOpen(constants.RepositoryPath(project))
	if err != nil {
		return err
	}

	prevRecord, _ := lo.Find(project.ExecuteRecords, func(item model.ProjectExecuteRecord) bool {
		return item.Branch == branchName
	})
	if err := migrateDatabaseSchema(project, branchName, repo, lo.Ternary(len(prevRecord.Hash) > 0, plumbing.NewHash(prevRecord.Hash), plumbing.ZeroHash)); err != nil {
		return err
	}

	return nil
}

func migrateDatabaseSchema(project model.Project, branchName string, repo *git.Repository, prevHash plumbing.Hash) error {
	ref, err := repo.Reference(plumbing.ReferenceName("refs/heads/"+branchName), true)
	if err != nil {
		return err
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return err
	}

	// 第一次提交或提交中修改了 database.json, 则需要应用新的数据库 schema
	databaseSchemaChanged := false
	// prevHash 为 plumbing.ZeroHash 说明是第一次提交
	if prevHash == plumbing.ZeroHash {
		databaseSchemaChanged = true
	} else {
		if prevCommit, err := repo.CommitObject(prevHash); err != nil {
			return err
		} else if diffs, err := git2.Diff(commit, prevCommit); err != nil {
			return err
		} else {
			databaseSchemaChanged = lo.ContainsBy(diffs, func(diff git2.DiffEntry) bool {
				return diff.FileName == "database.json"
			})
		}
	}
	if !databaseSchemaChanged {
		return nil
	}

	tree, err := commit.Tree()
	if err != nil {
		return err
	}
	databaseJson, err := tree.File("database.json")
	if err != nil {
		return err
	}
	// 校验 database.json
	if r, err := databaseJson.Reader(); err != nil {
		return err
	} else {
		if err := db_utils.ValidateDatabaseJson(r); err != nil {
			return err
		}
		defer r.Close()
	}

	if r, err := databaseJson.Reader(); err != nil {
		return err
	} else {
		databaseStruct := db_utils.Database{}
		if err := json.NewDecoder(r).Decode(&databaseStruct); err != nil {
			return err
		}

		if db, err := db_utils.OpenOrCreate(constants.DatabasePath(project, branchName)); err != nil {
			return err
		} else if err := db_utils.AutoMigrate(db, databaseStruct); err != nil {
			return err
		}
	}

	return nil
}
