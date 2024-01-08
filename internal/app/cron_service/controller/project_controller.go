package controller

import (
	"encoding/json"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/siaikin/home-dashboard/internal/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/model"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/service"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_errors"
)

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
	if err := c.ShouldBindJSON(&project); err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	}

	if count, err := service.CountProject(model.Project{Name: project.Name}); err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	} else if count > 0 {
		comfy_errors.ControllerUtils.RespondEntityAlreadyExistError(c, "project name already exists")
		return
	}

	affected, err := service.CreateOrUpdateProjects([]model.Project{project})
	if err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	}

	if err := initialRepository(project); err != nil {
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
	repoDir := filepath.Join(utils.WorkspaceDir(), "cron_service", "repos", project.Name)

	repo, err := git.PlainInit(repoDir, true)
	if err != nil {
		return err
	}
	if err := repo.CreateBranch(&config.Branch{
		Name: "master",
	}); err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	if err := worktree.Checkout(&git.CheckoutOptions{
		Branch: "master",
	}); err != nil {
		return err
	}

	type PackageJSON struct {
		Name            string            `json:"name"`
		Version         string            `json:"version"`
		Description     string            `json:"description"`
		Main            string            `json:"main"`
		Scripts         map[string]string `json:"scripts"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	packageJSONPath := filepath.Join(repoDir, "package.json")
	bytes, err := json.Marshal(PackageJSON{
		Name:            project.Name,
		Version:         "v0.0.0",
		Description:     project.Description,
		Main:            "main.js",
		Scripts:         nil,
		Dependencies:    nil,
		DevDependencies: nil,
	})
	if err != nil {
		return err

	}
	err = os.WriteFile(packageJSONPath, bytes, 0644)
	if err != nil {
		return err
	}

	if _, err := worktree.Add("."); err != nil {
		return err
	}

	hash, err := worktree.Commit("Initial Nodejs Repository", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "HOME Dashboard robot",
			Email: "abc1310054026@outlook.com",
			When:  time.Now(),
		},
	})

	if _, err := repo.CommitObject(hash); err != nil {
		return err
	}

	return nil

}
