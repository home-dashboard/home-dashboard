package monitor_service

import (
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_db"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
)

var logger = comfy_log.New("[monitor_service]")

var shortcutSectionModel = monitor_model.ShortcutSection{}

func CreateOrUpdateShortcutSections(sections []monitor_model.ShortcutSection) ([]monitor_model.ShortcutSection, error) {
	db := monitor_db.GetDB()

	affected := make([]monitor_model.ShortcutSection, len(sections))
	for i, section := range sections {
		model := db.Model(&shortcutSectionModel)

		if section.ID != 0 {
			result := model.Where(monitor_model.ShortcutSection{Model: monitor_model.Model{ID: section.ID}}).Assign(section).FirstOrCreate(&section)
			if result.Error != nil {
				return nil, result.Error
			}
		}

		if result := db.Save(&section); result.Error != nil {
			return nil, result.Error
		}
		affected[i] = section
	}

	return affected, nil
}

func DeleteShortcutSections(ids []uint) error {
	db := monitor_db.GetDB()

	result := db.Delete(&shortcutSectionModel, ids)

	return result.Error
}

func DeleteShortcutSectionItems(id uint, itemIds []uint) error {
	db := monitor_db.GetDB()

	items := make([]monitor_model.ShortcutItem, len(itemIds))
	for i, itemId := range itemIds {
		items[i] = monitor_model.ShortcutItem{Model: monitor_model.Model{ID: itemId}}
	}

	return db.Model(&monitor_model.ShortcutSection{Model: monitor_model.Model{
		ID: id,
	}}).Association("Items").Delete(items)
}

func ListShortcutSectionsByQuery(max int64, query monitor_model.ShortcutSection, preload []string) (*[]monitor_model.ShortcutSection, error) {
	return listShortcutSectionsWithPreload(max, query, preload)
}

func CountShortcutSection(query monitor_model.ShortcutSection) (int64, error) {
	db := monitor_db.GetDB()

	count := int64(0)
	result := db.Model(&shortcutSectionModel).Where(query).Count(&count)

	return count, result.Error
}

func listShortcutSectionsWithPreload(max int64, query monitor_model.ShortcutSection, preload []string) (*[]monitor_model.ShortcutSection, error) {
	db := monitor_db.GetDB()

	count, _ := CountShortcutSection(query)

	// 如果 max 大于记录总数或者 max 小于等于 0, 则返回所有记录.
	if max >= count || max <= 0 {
		max = count
	}

	sections := make([]monitor_model.ShortcutSection, max)

	model := db.Model(&shortcutSectionModel)
	for _, p := range preload {
		model = model.Preload(p)
	}

	result := model.Where(&query).Limit(int(max)).Find(&sections)

	return &sections, result.Error
}
