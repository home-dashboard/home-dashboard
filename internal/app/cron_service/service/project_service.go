package service

import (
	"reflect"
	"strings"

	"github.com/siaikin/home-dashboard/internal/app/cron_service/model"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_db"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"github.com/siaikin/home-dashboard/internal/pkg/database"
	"gorm.io/gorm/clause"
)

var projectModel = &model.Project{}

func CreateOrUpdateProjects(projects []model.Project) ([]model.Project, error) {
	db := monitor_db.GetDB()

	affected := make([]model.Project, len(projects))
	for i, project := range projects {
		tempModel := db.Model(&projectModel)

		if project.ID != 0 {
			result := tempModel.Where(model.Project{Model: monitor_model.Model{ID: project.ID}}).Assign(project).FirstOrCreate(&project)
			if result.Error != nil {
				return nil, result.Error
			}

		}

		if result := db.Save(&project); result.Error != nil {
			return nil, result.Error
		}
		affected[i] = project
	}

	return affected, nil
}

func DeleteProjects(ids []uint) error {
	db := monitor_db.GetDB()

	projects := make([]model.Project, len(ids))
	for i, id := range ids {
		projects[i] = model.Project{Model: monitor_model.Model{ID: id}}
	}
	if result := db.Delete(&projects); result.Error != nil {
		return result.Error
	}

	return nil
}

func ListProjectsByFuzzyQuery(query model.Project, likes []string) ([]model.Project, error) {
	return listProjects(query, likes)
}

func CountProject(query model.Project) (int64, error) {
	db := database.GetDB()

	count := int64(0)
	result := db.Model(&model.Project{}).Where(query).Count(&count)

	return count, result.Error
}

func listProjects(query model.Project, likes []string) ([]model.Project, error) {
	db := database.GetDB()

	projects := make([]model.Project, 0)

	model := db.Model(&projectModel)

	// 如果有 like 条件, 则将 like 条件从 query 中移除, 并将 like 条件添加到 likeClauses 中.
	queryValue := reflect.ValueOf(&query).Elem()
	likeClauses := make([]clause.Expression, len(likes))
	for _, l := range likes {
		field := queryValue.FieldByName(l)
		likeClauses = append(likeClauses, clause.Like{Column: l, Value: strings.Join([]string{"%", field.String(), "%"}, "")})
		field.Set(reflect.Zero(field.Type()))
	}
	model = model.Clauses(likeClauses...)

	result := model.Where(&query).Find(&projects)

	return projects, result.Error
}
