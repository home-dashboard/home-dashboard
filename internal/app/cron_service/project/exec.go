package project

import (
	"encoding/json"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/samber/lo"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/constants"
	git2 "github.com/siaikin/home-dashboard/internal/app/cron_service/git"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/model"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/project/db_utils"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/project/runner"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"path"
	"reflect"
)

func Exec(project model.Project, branchName string) error {
	repo, err := git.PlainOpen(constants.RepositoryPath(project))
	if err != nil {
		return fmt.Errorf("git.PlainOpen failed, %w", err)
	}

	prevRecord, ok := lo.Find(project.ExecuteRecords, func(item model.ProjectExecuteRecord) bool {
		return item.Branch == branchName
	})

	db, err := db_utils.OpenOrCreate(constants.DatabasePath(project, branchName))
	if err != nil {
		return fmt.Errorf("db_utils.OpenOrCreate failed, %w", err)
	}

	tables, err := migrateDatabaseSchema(db, branchName, repo, lo.Ternary(ok, plumbing.NewHash(prevRecord.Hash), plumbing.ZeroHash))
	if err != nil {
		return fmt.Errorf("migrateDatabaseSchema failed, %w", err)
	}

	nodejsRunner := runner.NewNodejsRunner(project, constants.NodejsPath)
	eg := new(errgroup.Group)

	eg.Go(func() error {
		return nodejsRunner.Run(branchName)
	})
	eg.Go(func() error {
	loop:
		for {
			select {
			case data, ok := <-nodejsRunner.C:
				if !ok {
					break loop
				}

				tableName := data.Table
				tableSchema, ok := tables[tableName]
				if !ok {
					fmt.Printf("table not found, %s", tableName)
					continue
				}

				for _, row := range data.Rows {
					tableSchemaCopy := reflect.New(reflect.TypeOf(tableSchema).Elem()).Interface()
					jsonData, _ := json.Marshal(row)
					_ = json.Unmarshal(jsonData, &tableSchemaCopy)

					result := db.Table(tableName).Create(tableSchemaCopy)
					if result.Error != nil {
						fmt.Printf("create row failed, %w", result.Error)
					}
				}
			}
		}
		return nil
	})

	return eg.Wait()
}

// migrateDatabaseSchema 检查是否需要应用新的数据库 schema
func migrateDatabaseSchema(db *gorm.DB, branchName string, repo *git.Repository, prevHash plumbing.Hash) (map[string]any, error) {
	ref, err := repo.Reference(plumbing.ReferenceName("refs/heads/"+branchName), true)
	if err != nil {
		return nil, err
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}

	// 第一次提交或提交中修改了 database.json, 则需要应用新的数据库 schema
	databaseSchemaChanged := false
	// prevHash 为 plumbing.ZeroHash 说明是第一次提交
	if prevHash == plumbing.ZeroHash {
		databaseSchemaChanged = true
	} else {
		prevCommit, err := repo.CommitObject(prevHash)
		if err != nil {
			return nil, err
		}
		diffs, err := git2.Diff(commit, prevCommit)
		if err != nil {
			return nil, err
		}
		databaseSchemaChanged = lo.ContainsBy(diffs, func(diff git2.DiffEntry) bool {
			return diff.FileName == "database.json"
		})
	}
	if !databaseSchemaChanged {
		return nil, err
	}

	databaseJsonString, err := extractDatabaseJsonContent(commit)
	if err != nil {
		return nil, err
	}

	err = db_utils.ValidateDatabaseJson(databaseJsonString)
	if err != nil {
		return nil, err
	}

	tables, err := parseDatabaseSchema(db, databaseJsonString)
	if err != nil {
		return nil, err
	}

	for name, columns := range tables {
		if err := db.Table(name).AutoMigrate(columns); err != nil {
			return nil, err
		}
	}

	return tables, nil
}

// parseDatabaseSchema 解析 git 仓库中 branchName 分支的 database.json 并返回.
func parseDatabaseSchema(db *gorm.DB, str string) (map[string]any, error) {
	databaseStruct := db_utils.Database{}
	if err := json.Unmarshal([]byte(str), &databaseStruct); err != nil {
		return nil, err
	}

	result := make(map[string]any, len(databaseStruct.Tables))
	for _, table := range databaseStruct.Tables {
		result[table.Name] = db_utils.TableStructToGormModel(db, table)
	}

	return result, nil
}

func extractDatabaseJsonContent(commit *object.Commit) (string, error) {
	tree, err := commit.Tree()
	if err != nil {
		return "", err
	}
	databaseJson, err := tree.File(path.Join(constants.ProjectRepoSpecificDir, "database.json"))
	if err != nil {
		return "", err
	}

	return databaseJson.Contents()
}
