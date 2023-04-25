package monitor_model

import "gorm.io/plugin/soft_delete"

type Model struct {
	// Id 允许读和创建, 不可写
	ID uint `gorm:"<-:create; primarykey" json:"id"`
	// CreatedAt 创建时间, 允许读和创建, 不可写
	CreatedAt int64 `gorm:"<-:create; autoUpdateTime:milli" json:"createdAt"`
	// UpdatedAt 更新时间, 可读可写
	UpdatedAt int64 `gorm:"autoUpdateTime:milli" json:"updatedAt"`
	// DeletedAt 删除时间, 可读可写
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli" json:"deletedAt"`
}
