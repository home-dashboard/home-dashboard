package server_monitor

import (
	"context"
	"errors"
	"github.com/jinzhu/copier"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_db"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_process_realtime"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_realtime"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_service"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/user_notification"
	"github.com/siaikin/home-dashboard/internal/pkg/authority"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"github.com/siaikin/home-dashboard/internal/pkg/configuration"
	"gorm.io/gorm"
	"time"
)

var logger = comfy_log.New("[server_monitor]")

var ctx context.Context
var cancel context.CancelFunc

func Initial(db *gorm.DB) {
	monitor_db.Initial(db)

	if err := initialDeviceTable(); err != nil {
		panic(err)
	}

	if err := generateAdministratorUser(); err != nil {
		panic(err)
	}

	if err := storeLatestConfiguration(); err != nil {
		panic(err)
	}
}

func Start(port uint) {
	ctx, cancel = context.WithCancel(context.Background())

	monitor_realtime.Loop(ctx, time.Second)
	monitor_process_realtime.Loop(ctx, time.Second)
	user_notification.StartListenUserNotificationNotify(ctx)

	startServer(port, configuration.Get().ServerMonitor.Development.Enable)
}

func Stop() {
	cancel()

	stopServer(5 * time.Second)
}

func initialDeviceTable() error {
	db := monitor_db.GetDB()

	systemStat := monitor_realtime.GetCachedSystemRealtimeStat()

	for _, v := range systemStat.Network {
		networkInfo := monitor_model.StoredSystemNetworkAdapterInfo{
			SystemNetworkAdapterInfo: monitor_realtime.SystemNetworkAdapterInfo{
				Index:       uint32(v.InterfaceStat.Index),
				Type:        v.InterfaceStat.Type,
				Description: v.InterfaceStat.Description,
			},
		}

		db.Where(map[string]interface{}{"Index": networkInfo.Index}).Attrs(networkInfo).FirstOrCreate(&networkInfo)
	}

	for _, v := range systemStat.Cpu {
		cpuInfo := monitor_model.StoredSystemCpuInfo{}
		if err := copier.Copy(&cpuInfo, &v.InfoStat); err != nil {
			logger.Info("copy failed, %s\n", err)
			return err
		}

		db.Where(map[string]interface{}{"CPU": cpuInfo.CPU}).Attrs(cpuInfo).FirstOrCreate(&cpuInfo)
	}

	for _, v := range systemStat.Disk {
		diskInfo := monitor_model.StoredSystemDiskInfo{}
		if err := copier.Copy(&diskInfo, &v.PartitionStat); err != nil {
			logger.Info("copy failed, %s\n", err)
			return err
		}

		db.Where(map[string]interface{}{"Mountpoint": diskInfo.Mountpoint}).Attrs(diskInfo).FirstOrCreate(&diskInfo)
	}

	return nil
}

// 从配置文件中的管理员配置中生成管理员账号.
func generateAdministratorUser() error {
	administrator := configuration.Get().ServerMonitor.Administrator

	user, err := monitor_service.GetUser(monitor_model.User{Role: monitor_model.RoleAdministrator})

	// 管理员账号密码不一致则删除账号并重新创建管理员账号
	if err == nil && (user.Username != administrator.Username || user.Password != administrator.Password) {
		if err := monitor_service.DeleteUser(*user); err != nil {
			return err
		}

		// 给 err 赋值以进入账号创建分支
		err = errors.New("")
	}

	if err != nil {
		adminUser := monitor_model.User{
			User: authority.User{
				Username: administrator.Username,
				Password: administrator.Password,
			},
			Role: monitor_model.RoleAdministrator,
		}

		if err := monitor_service.CreateUser(adminUser); err != nil {
			return err
		}
	}

	return nil
}

// 存储最新的配置信息到数据库中
func storeLatestConfiguration() error {
	currentConfig := configuration.Get()

	if config, err := monitor_service.LatestConfiguration(); err != nil {
		return err
	} else if config != nil && config.Configuration.ModificationTime == currentConfig.ModificationTime { // 记录条数为 0 或配置文件未修改时跳过
		return nil
	}

	if err := monitor_service.CreateConfiguration(monitor_model.StoredConfiguration{Configuration: *configuration.Get()}); err != nil {
		return err
	}

	return nil
}
