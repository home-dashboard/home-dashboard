package monitor_model

import (
	"github.com/siaikin/home-dashboard/internal/pkg/configuration"
)

type StoredConfiguration struct {
	Model
	Configuration configuration.Configuration `gorm:"serializer:gob" json:"configuration"`
}
