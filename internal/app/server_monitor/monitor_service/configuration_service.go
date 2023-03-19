package monitor_service

import (
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"github.com/siaikin/home-dashboard/internal/pkg/database"
)

var configurationModel = monitor_model.StoredConfiguration{}

// LatestConfiguration 获取最新的配置记录, 如表中记录为空, 并不会返回 error 而是返回的两个参数都为 nil.
func LatestConfiguration() (*monitor_model.StoredConfiguration, error) {
	list, err := LatestNConfiguration(1)

	if len(*list) <= 0 {
		return nil, err
	} else {
		return &(*list)[0], err
	}
}

// LatestNConfiguration 获取前 max 个配置记录, 根据插入顺序排序. 当记录条数小于 max 时, 返回的结果数组长度为实际记录条数的长度.
func LatestNConfiguration(max int64) (*[]monitor_model.StoredConfiguration, error) {
	db := database.GetDB()

	count, _ := CountConfiguration()

	if max > *count {
		max = *count
	}

	configs := make([]monitor_model.StoredConfiguration, max)
	result := db.Offset(int(*count - max)).Limit(int(max)).Model(&configurationModel).Find(&configs)

	return &configs, result.Error
}

// CreateConfiguration 插入一条记录
func CreateConfiguration(config monitor_model.StoredConfiguration) error {
	db := database.GetDB()

	result := db.Create(&config)

	return result.Error
}

// DeleteConfiguration 删除一条记录
func DeleteConfiguration(config monitor_model.StoredConfiguration) error {
	db := database.GetDB()

	result := db.Delete(&config)

	return result.Error
}

// CountConfiguration 获取记录的总条数
func CountConfiguration() (*int64, error) {
	db := database.GetDB()

	count := int64(0)
	result := db.Model(&configurationModel).Count(&count)

	return &count, result.Error
}
