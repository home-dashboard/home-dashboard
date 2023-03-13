package server_monitor

import (
	"errors"
	"github.com/jinzhu/copier"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_db"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_process_realtime"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_realtime"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_service"
	"github.com/siaikin/home-dashboard/internal/pkg/configuration"
	"log"
	"time"
)

func Initial() {
	if err := initialDeviceTable(); err != nil {
		panic(err)
	}

	if err := generateAdministratorUser(); err != nil {
		panic(err)
	}
}

func Start(port uint) {
	monitor_realtime.StartSystemRealtimeStatLoop(time.Second)
	monitor_process_realtime.StartRealtimeLoop(time.Second)

	startServer(port, configuration.Config.ServerMonitor.Development.Enable)
}

func Stop() {
	monitor_realtime.StopSystemRealtimeStatLoop()
	monitor_process_realtime.StopRealtimeLoop()

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
			log.Printf("cpoy failed, %s", err)
			return err
		}

		db.Where(map[string]interface{}{"CPU": cpuInfo.CPU}).Attrs(cpuInfo).FirstOrCreate(&cpuInfo)
	}

	for _, v := range systemStat.Disk {
		diskInfo := monitor_model.StoredSystemDiskInfo{}
		if err := copier.Copy(&diskInfo, &v.PartitionStat); err != nil {
			log.Printf("cpoy failed, %s", err)
			return err
		}

		db.Where(map[string]interface{}{"Mountpoint": diskInfo.Mountpoint}).Attrs(diskInfo).FirstOrCreate(&diskInfo)
	}

	return nil
}

// 从配置文件中的管理员配置中生成管理员账号.
func generateAdministratorUser() error {
	administrator := configuration.Config.ServerMonitor.Administrator

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
			Username: administrator.Username,
			Password: administrator.Password,
			Role:     monitor_model.RoleAdministrator,
		}

		if err := monitor_service.CreateUser(adminUser); err != nil {
			return err
		}
	}

	return nil
}
