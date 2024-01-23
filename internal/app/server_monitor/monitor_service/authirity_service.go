package monitor_service

import (
	"github.com/go-errors/errors"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_db"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"github.com/siaikin/home-dashboard/internal/pkg/authority"
)

var userModel = monitor_model.User{}

var (
	ErrorNotFound = errors.New("not found")
)

func GetUserByName(username string) (monitor_model.User, error) {
	return GetUser(monitor_model.User{User: authority.User{Username: username}})
}

func GetUser(query monitor_model.User) (monitor_model.User, error) {
	db := monitor_db.GetDB()

	count, _ := CountUser(query)

	user := monitor_model.User{}
	if count > 0 {
		result := db.Model(&userModel).Where(&query).First(&user)
		return user, result.Error
	} else {
		return user, ErrorNotFound
	}
}

func CreateUser(user monitor_model.User) error {
	db := monitor_db.GetDB()

	result := db.Create(&user)

	return result.Error
}

func UpdateUser(user monitor_model.User) error {
	db := monitor_db.GetDB()

	result := db.Save(&user)

	return result.Error
}

func DeleteUser(user monitor_model.User) error {
	db := monitor_db.GetDB()

	result := db.Delete(&user)

	return result.Error
}

// CountUser 获取 User 记录的总条数
func CountUser(user monitor_model.User) (int64, error) {
	db := monitor_db.GetDB()

	count := int64(0)
	result := db.Model(&userModel).Where(user).Count(&count)

	return count, result.Error
}
