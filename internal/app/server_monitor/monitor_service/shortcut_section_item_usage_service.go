package monitor_service

import (
	"github.com/samber/lo"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_db"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
)

var shortcutSectionItemUsageModel = monitor_model.ShortcutSectionItemUsage{}

func CreateOrUpdateShortcutSectionItemUsages(usages []monitor_model.ShortcutSectionItemUsage) ([]monitor_model.ShortcutSectionItemUsage, error) {
	db := monitor_db.GetDB()

	usages = lo.Filter(usages, func(usage monitor_model.ShortcutSectionItemUsage, index int) bool {
		return usage.SectionId > 0 && usage.ItemId > 0
	})

	// 如果 usages 中有记录在数据库中不存在, 先创建这些记录.
	createdUsages := lo.Filter(usages, func(usage monitor_model.ShortcutSectionItemUsage, index int) bool {
		count, err := CountShortcutSectionItemUsage(monitor_model.ShortcutSectionItemUsage{SectionId: usage.SectionId, ItemId: usage.ItemId})
		return err == nil && count <= 0
	})
	for _, usage := range createdUsages {
		if result := db.Create(&usage); result.Error != nil {
			return nil, result.Error
		}
		// 把 monitor_model.ShortcutSectionItemUsage 关联到 monitor_model.ShortcutItem 的 Usages 中.
		if err := db.Model(&shortcutSectionItemUsageModel).Where(monitor_model.Model{ID: usage.ItemId}).Association("Usages").Append(&usage); err != nil {
			return nil, err
		}
	}

	affected := make([]monitor_model.ShortcutSectionItemUsage, len(usages))
	for i, usage := range usages {
		model := db.Model(&shortcutSectionItemUsageModel)

		storedUsage := monitor_model.ShortcutSectionItemUsage{}
		result := model.Where(monitor_model.ShortcutSectionItemUsage{SectionId: usage.SectionId, ItemId: usage.ItemId}).First(&storedUsage)
		if result.Error != nil {
			return nil, result.Error
		}

		usage.ClickCount += storedUsage.ClickCount

		if result := db.Model(usage).Update("ClickCount", usage.ClickCount); result.Error != nil {
			return nil, result.Error
		}
		affected[i] = usage
	}

	return affected, nil
}

func DeleteShortcutSectionItemUsages(ids []uint) error {
	db := monitor_db.GetDB()

	result := db.Delete(&shortcutSectionItemUsageModel, ids)

	return result.Error
}

func ListShortcutSectionItemUsagesByQuery(query monitor_model.ShortcutSectionItemUsage, preload []string) (*[]monitor_model.ShortcutSectionItemUsage, error) {
	return listShortcutSectionItemUsagesWithPreload(query, preload)
}

func CountShortcutSectionItemUsage(query monitor_model.ShortcutSectionItemUsage) (int64, error) {
	db := monitor_db.GetDB()

	count := int64(0)
	result := db.Model(&shortcutSectionItemUsageModel).Where(query).Count(&count)

	return count, result.Error
}

func listShortcutSectionItemUsagesWithPreload(query monitor_model.ShortcutSectionItemUsage, preload []string) (*[]monitor_model.ShortcutSectionItemUsage, error) {
	db := monitor_db.GetDB()

	count, _ := CountShortcutSectionItemUsage(query)
	usages := make([]monitor_model.ShortcutSectionItemUsage, count)

	model := db.Model(&shortcutSectionItemUsageModel)
	for _, p := range preload {
		model = model.Preload(p)
	}

	result := model.Where(&query).Find(&usages)

	return &usages, result.Error
}
