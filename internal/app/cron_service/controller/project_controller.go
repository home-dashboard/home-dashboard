package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/constants"
	git2 "github.com/siaikin/home-dashboard/internal/app/cron_service/git"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/model"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/project"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/service"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_errors"
	"net/http"
	"regexp"
	"strconv"
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
	var proj model.Project
	if err := c.BindJSON(&proj); err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	} else if !repositoryNameRegexp.MatchString(proj.Name) {
		comfy_errors.ControllerUtils.RespondEntityValidationError(c, "invalid project name")
		return
	}

	if count, err := service.CountProject(model.Project{Name: proj.Name}); err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	} else if count > 0 {
		comfy_errors.ControllerUtils.RespondEntityAlreadyExistError(c, "project name already exists")
		return
	}

	logger.Info("initial repository for project %s...", proj.Name)
	if err := initialRepository(proj); err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	}
	logger.Info("initial repository for project %s done", proj.Name)

	affected, err := service.CreateOrUpdateProjects([]model.Project{proj})
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

func initialRepository(proj model.Project) error {
	repoDir := constants.RepositoryPath(proj)

	return git2.CloneBareFromGitHubTemplate(repoDir, constants.ProjectTemplateUrl(proj.RunnerType))
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
