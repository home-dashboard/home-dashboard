package monitor_model

import "gorm.io/plugin/soft_delete"

type Model struct {
	ID        uint                  `gorm:"primarykey" json:"id"`
	CreatedAt int64                 `gorm:"autoUpdateTime:milli" json:"createdAt"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli" json:"updatedAt"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli" json:"deletedAt"`
}
