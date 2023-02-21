package monitor_service

import (
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_db"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
)

func GetUserByName(username string) (*monitor_model.User, error) {
	return GetUser(monitor_model.User{Username: username})
}

func GetUser(user monitor_model.User) (*monitor_model.User, error) {

	db := monitor_db.GetDB()

	result := db.Where(&user).First(&user)

	return &user, result.Error
}

func CreateUser(user monitor_model.User) error {
	db := monitor_db.GetDB()

	result := db.Create(&user)

	return result.Error
}

func DeleteUser(user monitor_model.User) error {
	db := monitor_db.GetDB()

	result := db.Delete(&user)

	return result.Error
}
