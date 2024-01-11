package runner

import (
	"github.com/siaikin/home-dashboard/internal/app/cron_service/model"
)

type Runner interface {
	// Run 执行 Project
	Run(project model.Project) error
}
