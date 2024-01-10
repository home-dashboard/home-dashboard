package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/constants"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/constants/templates"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/model"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/project"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/service"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_errors"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var repositoryNameRegexp = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// CreateProject 创建 Project
// @Summary CreateProject
// @Description CreateProject
// @Tags CreateProject
// @Produce json
// @Param project body Project true "Project"
// @Success 200 {object} string
// @Router project/create [post]
func CreateProject(c *gin.Context) {
	var project model.Project
	if err := c.BindJSON(&project); err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	} else if !repositoryNameRegexp.MatchString(project.Name) {
		comfy_errors.ControllerUtils.RespondEntityValidationError(c, "invalid project name")
		return
	}

	if count, err := service.CountProject(model.Project{Name: project.Name}); err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	} else if count > 0 {
		comfy_errors.ControllerUtils.RespondEntityAlreadyExistError(c, "project name already exists")
		return
	}

	if err := initialRepository(project); err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	}

	affected, err := service.CreateOrUpdateProjects([]model.Project{project})
	if err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, affected[0])
}

// ListProjects 列出 Project
// @Summary ListProjects
// @Description ListProjects
// @Tags ListProjects
// @Produce json
// @Param project body Project true "Project"
// @Success 200 {object} string
// @Router project/list [get]
func ListProjects(c *gin.Context) {
	var project model.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	}

	if projects, err := service.ListProjectsByFuzzyQuery(project, []string{"name", "display_name"}); err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	} else {
		c.JSON(http.StatusOK, projects)
	}
}

// UpdateProject 更新 Project 信息
// @Summary UpdateProject
// @Description UpdateProject
// @Tags UpdateProject
// @Produce json
// @Param project body Project true "Project"
// @Success 200 {object} string
// @Router project/update/{id} [put]
func UpdateProject(c *gin.Context) {
	var project model.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	} else if !repositoryNameRegexp.MatchString(project.Name) {
		comfy_errors.ControllerUtils.RespondEntityValidationError(c, "invalid project name")
		return
	} else if ID, err := strconv.ParseUint(c.Param("id"), 10, 0); err != nil {
		comfy_errors.ControllerUtils.RespondEntityValidationError(c, "invalid id")
		return
	} else {
		project.ID = uint(ID)
	}

	if updated, err := service.CreateOrUpdateProjects([]model.Project{project}); err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	} else {
		c.JSON(http.StatusOK, updated[0])
	}
}

// DeleteProject 删除 Project
// @Summary DeleteProject
// @Description DeleteProject
// @Tags DeleteProject
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} string
// @Router project/delete/{id} [delete]
func DeleteProject(c *gin.Context) {
	if ID, err := strconv.ParseUint(c.Param("id"), 10, 0); err != nil {
		comfy_errors.ControllerUtils.RespondEntityValidationError(c, "invalid id")
		return
	} else if err := service.DeleteProjects([]uint{uint(ID)}); err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	} else {
		c.Status(http.StatusOK)
	}
}

func initialRepository(project model.Project) error {
	repoDir := constants.RepositoryPath(project)

	repo, err := git.PlainInitWithOptions(repoDir, &git.PlainInitOptions{
		InitOptions: git.InitOptions{
			DefaultBranch: "refs/heads/master",
		},
		Bare: true,
	})
	if err != nil {
		return err
	}

	tree := object.Tree{
		Entries: []object.TreeEntry{},
	}

	createRepositoryFile := func(fileName string, data any) error {
		fileObject := repo.Storer.NewEncodedObject()
		fileObject.SetType(plumbing.BlobObject)
		w, err := fileObject.Writer()
		if err != nil {
			return err
		}

		if err := templates.ExecuteTemplate(fileName, data, w); err != nil {
			return err
		}
		if err := w.Close(); err != nil {
			return err
		}

		fileHash, err := repo.Storer.SetEncodedObject(fileObject)
		if err != nil {
			return err
		}

		treeEntry := object.TreeEntry{
			Name: fileName,
			Mode: filemode.Regular,
			Hash: fileHash,
		}

		tree.Entries = append(tree.Entries, treeEntry)

		return nil
	}

	if err := createRepositoryFile(".gitattributes", nil); err != nil {
		return err
	}
	if err := createRepositoryFile(".gitignore", nil); err != nil {
		return err
	}
	if err := createRepositoryFile("README.md", nil); err != nil {
		return err
	}
	if err := createRepositoryFile("package.json", templates.PackageJSONData{
		Name:            project.Name,
		Version:         "v0.0.0",
		Description:     project.Description,
		Main:            "main.js",
		Scripts:         nil,
		Dependencies:    nil,
		DevDependencies: nil,
	}); err != nil {
		return err
	}
	if err := createRepositoryFile("database.json", map[string]any{"SchemaUrl": "https://gist.githubusercontent.com/siaikin/e14e83015bb93b651ccbb1b060397673/raw/8acdd63c266d23ac800e9b570b0edd1145ce28ae/database_schema.json"}); err != nil {
		return err
	}

	treeObject := repo.Storer.NewEncodedObject()
	err = tree.Encode(treeObject)
	if err != nil {
		return err
	}

	treeHash, err := repo.Storer.SetEncodedObject(treeObject)
	if err != nil {
		return err
	}

	newCommit := object.Commit{
		Author:    object.Signature{Name: "siaikin", Email: "abc1310054026@outlook.com", When: time.Date(2024, 1, 9, 5, 38, 0, 0, time.UTC)},
		Committer: object.Signature{Name: "HOME Dashboard robot", Email: "abc1310054026@outlook.com", When: time.Now()},
		Message:   "Initial commit",
		TreeHash:  treeHash,
	}

	commitObject := repo.Storer.NewEncodedObject()
	err = newCommit.Encode(commitObject)
	if err != nil {
		return err
	}

	commitHash, err := repo.Storer.SetEncodedObject(commitObject)
	if err != nil {
		return err
	}

	// Now, point the "main" branch to the newly-created commit

	ref := plumbing.NewHashReference("refs/heads/master", commitHash)
	err = repo.Storer.SetReference(ref)
	if err != nil {
		return err
	}

	return nil

}

// RunProject 运行 Project
// @Summary RunProject
// @Description RunProject
// @Tags RunProject
// @Produce json
// @Param project path string true "Project Name"
// @Param branch query string true "Branch Name"
// @Success 200 {object} string
// @Router project/run/{project} [post]
func RunProject(c *gin.Context) {
	projectName := c.Param("project")
	branchName := c.Query("branch")

	if err := project.Exec(model.Project{
		Name: projectName,
	}, branchName); err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	}

	c.Status(http.StatusOK)
}
