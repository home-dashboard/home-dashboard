package monitor_service

import (
	"errors"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"github.com/siaikin/home-dashboard/internal/pkg/database"
)

var userModel = monitor_model.User{}

func GetUserByName(username string) (*monitor_model.User, error) {
	return GetUser(monitor_model.User{Username: username})
}

func GetUser(user monitor_model.User) (*monitor_model.User, error) {

	db := database.GetDB()

	count, _ := CountUser()

	if *count > 0 {
		result := db.Where(&user).First(&user)
		return &user, result.Error
	} else {
		return nil, errors.New("not found")
	}
}

func CreateUser(user monitor_model.User) error {
	db := database.GetDB()

	result := db.Create(&user)

	return result.Error
}

func DeleteUser(user monitor_model.User) error {
	db := database.GetDB()

	result := db.Delete(&user)

	return result.Error
}

// CountUser 获取 User 记录的总条数
func CountUser() (*int64, error) {
	db := database.GetDB()

	count := int64(0)
	result := db.Model(&userModel).Count(&count)

	return &count, result.Error
}
