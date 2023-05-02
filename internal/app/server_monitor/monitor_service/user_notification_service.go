package monitor_service

import (
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_db"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"github.com/siaikin/home-dashboard/internal/pkg/notification"
)

var notificationModel = monitor_model.StoredNotification{}

// LatestNNotification 获取前 max 个通知, 根据插入顺序排序. 当记录条数小于 max 时, 返回的结果数组长度为实际记录条数的长度.
// 如果 onlyUnread 为 true, 则只返回未读的通知. 否则返回所有通知.
func LatestNNotification(max int64, onlyUnread bool) (*[]monitor_model.StoredNotification, error) {
	return listNNotification(max, notification.UserNotification{Unread: onlyUnread})
}
func listNNotification(max int64, query notification.UserNotification) (*[]monitor_model.StoredNotification, error) {
	db := monitor_db.GetDB()

	count, _ := CountNotification()

	// 如果 max 大于记录总数或者 max 小于等于 0, 则返回所有记录.
	if max >= *count || max <= 0 {
		max = *count
	}

	configs := make([]monitor_model.StoredNotification, max)

	result := db.Order("origin_create_at desc").Where(&monitor_model.StoredNotification{UserNotification: query}).Limit(int(max)).Model(&notificationModel).Find(&configs)

	return &configs, result.Error
}

// LatestNNotificationByQuery 获取前 max 个通知, 根据插入顺序排序. 当记录条数小于 max 时, 返回的结果数组长度为实际记录条数的长度. 根据 query 进行过滤.
func LatestNNotificationByQuery(max int64, query notification.UserNotification) (*[]monitor_model.StoredNotification, error) {
	return listNNotification(max, query)
}

// GetNotification 根据给定的 id 获取一条通知记录.
func GetNotification(id uint) (*monitor_model.StoredNotification, error) {
	db := monitor_db.GetDB()

	config := monitor_model.StoredNotification{}
	result := db.Model(&notificationModel).First(&config, id)

	return &config, result.Error
}

// CreateNotification 插入一条通知记录.
func CreateNotification(_notification monitor_model.StoredNotification) error {
	db := monitor_db.GetDB()

	// 如果已存在相同的记录, 则不插入.
	count := int64(0)
	db.Model(&notificationModel).Where(&monitor_model.StoredNotification{UserNotification: notification.UserNotification{UniqueId: _notification.UniqueId}}).Count(&count)

	if count > 0 {
		return nil
	}

	result := db.Create(&_notification)

	return result.Error
}

// CreateOrUpdateNotification 插入或更新单条通知记录.
func CreateOrUpdateNotification(_notification monitor_model.StoredNotification) error {
	return CreateOrUpdateNotifications([]monitor_model.StoredNotification{_notification})
}

// CreateOrUpdateNotifications 插入或更新多条通知记录.
func CreateOrUpdateNotifications(notifications []monitor_model.StoredNotification) error {
	db := monitor_db.GetDB()

	for _, _notification := range notifications {
		result := db.Model(&notificationModel).Where(monitor_model.StoredNotification{UserNotification: notification.UserNotification{UniqueId: _notification.UniqueId}}).Attrs(_notification).FirstOrCreate(&_notification)
		if result.Error != nil {
			return result.Error
		}
		// 受影响的行数大于 0 说明记录已存在, 不需要更新.
		if result.RowsAffected > 0 {
			continue
		}

		if result = db.Save(&_notification); result.Error != nil {
			return result.Error
		}
	}

	return nil
}

// MarkNotificationAsRead 标记一条通知为已读
func MarkNotificationAsRead(id uint) error {
	return MarkNotificationsAsRead([]uint{id})
}

// MarkNotificationsAsRead 标记多条通知为已读
func MarkNotificationsAsRead(ids []uint) error {
	db := monitor_db.GetDB()

	result := db.Model(&notificationModel).Where("id IN ?", ids).Update("unread", false)

	return result.Error
}

// SyncNotificationsReadStateByUniqueIds 同步用户通知的已读状态. 从数据库中查询所有 origin 下的未读通知,
// 如果通知的 uniqueId 不在 unreadUniqueIds 中, 则将其标记为已读.
// 处理完成后将返回一个被标记为已读的通知 id 列表.
func SyncNotificationsReadStateByUniqueIds(unreadUniqueIds []string, origin string) ([]uint, error) {
	unreadUniqueIdMap := make(map[string]bool)
	for _, uniqueId := range unreadUniqueIds {
		unreadUniqueIdMap[uniqueId] = true
	}

	db := monitor_db.GetDB()

	notifications, err := LatestNNotificationByQuery(0, notification.UserNotification{Unread: true, Origin: origin})
	if err != nil {
		return nil, err
	}

	readIds := make([]uint, 0)
	// 遍历数据库对应来源的所有未读通知, 如果通知的 uniqueId 不在 unreadUniqueIds 中, 则将其加入到已读 id 列表
	for _, unreadNotification := range *notifications {
		// 跳过已在 unreadUniqueIds 中的通知
		if _, ok := unreadUniqueIdMap[unreadNotification.UniqueId]; ok {
			continue
		}

		readIds = append(readIds, unreadNotification.ID)
	}

	if result := db.Model(&notificationModel).Where("id IN ?", readIds).Update("unread", false); result.Error != nil {
		return nil, result.Error
	}

	return readIds, nil
}

// MarkAllNotificationAsRead 标记所有未读通知为已读
func MarkAllNotificationAsRead() error {
	db := monitor_db.GetDB()

	result := db.Model(&notificationModel).Where(map[string]any{"unread": true}).Update("unread", true)

	return result.Error
}

// DeleteNotification 删除一条通知记录
func DeleteNotification(id uint) error {
	db := monitor_db.GetDB()

	result := db.Delete(&notificationModel, id)

	return result.Error
}

// CountNotification 获取记录的总条数
func CountNotification() (*int64, error) {
	db := monitor_db.GetDB()

	count := int64(0)
	result := db.Model(&notificationModel).Count(&count)

	return &count, result.Error
}
