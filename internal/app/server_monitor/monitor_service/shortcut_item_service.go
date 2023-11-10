package monitor_service

import (
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_db"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"gorm.io/gorm/clause"
	"reflect"
	"strings"
)

var shortcutItemModel = monitor_model.ShortcutItem{}

func CreateOrUpdateShortcutItems(items []monitor_model.ShortcutItem) ([]monitor_model.ShortcutItem, error) {
	db := monitor_db.GetDB()

	affected := make([]monitor_model.ShortcutItem, len(items))
	for i, item := range items {
		model := db.Model(&shortcutItemModel)

		if item.ID != 0 {
			result := model.Where(monitor_model.ShortcutItem{Model: monitor_model.Model{ID: item.ID}}).Assign(item).FirstOrCreate(&item)
			if result.Error != nil {
				return nil, result.Error
			}
		}

		if result := db.Save(&item); result.Error != nil {
			return nil, result.Error
		}
		affected[i] = item
	}

	return affected, nil
}

// DeleteShortcutItems 删除 monitor_model.ShortcutItem.
// 如果快捷方式已被 monitor_model.ShortcutSection 引用, 则也会删除 monitor_model.ShortcutItem 与 monitor_model.ShortcutSection 的关联关系.
func DeleteShortcutItems(ids []uint) error {
	db := monitor_db.GetDB()

	items := make([]monitor_model.ShortcutItem, len(ids))
	for i, id := range ids {
		items[i] = monitor_model.ShortcutItem{Model: monitor_model.Model{ID: id}}
	}
	if result := db.Select("Sections").Delete(&items); result.Error != nil {
		return result.Error
	}

	return nil
}

func ListShortcutItemsByQuery(query monitor_model.ShortcutItem, preload []string) (*[]monitor_model.ShortcutItem, error) {
	return listShortcutItemsWithPreload(query, preload, []string{})
}

func ListShortcutItemsByFuzzyQuery(query monitor_model.ShortcutItem, likes []string, preload []string) (*[]monitor_model.ShortcutItem, error) {
	return listShortcutItemsWithPreload(query, preload, likes)
}

func CountShortcutItem(query monitor_model.ShortcutItem) (int64, error) {
	db := monitor_db.GetDB()

	count := int64(0)
	result := db.Model(&shortcutItemModel).Where(query).Count(&count)

	return count, result.Error
}

func listShortcutItemsWithPreload(query monitor_model.ShortcutItem, preload []string, likes []string) (*[]monitor_model.ShortcutItem, error) {
	db := monitor_db.GetDB()

	items := make([]monitor_model.ShortcutItem, 0)

	model := db.Model(&shortcutItemModel)
	for _, p := range preload {
		model = model.Preload(p)
	}

	// 如果有 like 条件, 则将 like 条件从 query 中移除, 并将 like 条件添加到 likeClauses 中.
	queryValue := reflect.ValueOf(&query).Elem()
	likeClauses := make([]clause.Expression, len(likes))
	for _, l := range likes {
		field := queryValue.FieldByName(l)
		likeClauses = append(likeClauses, clause.Like{Column: l, Value: strings.Join([]string{"%", field.String(), "%"}, "")})
		field.Set(reflect.Zero(field.Type()))
	}
	model = model.Clauses(likeClauses...)

	result := model.Where(&query).Find(&items)

	return &items, result.Error
}
