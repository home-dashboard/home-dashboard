package server_monitor

import (
	"context"
	"github.com/go-errors/errors"
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
	"net"
	"net/http"
	"time"
)

var logger = comfy_log.New("[server_monitor]")

func Initial(db *gorm.DB) error {
	if err := monitor_db.Initial(db); err != nil {
		return err
	}

	logger.Info("initial device table...\n")
	if err := initialDeviceTable(); err != nil {
		return err
	}

	logger.Info("generate administrator user...\n")
	if err := generateAdministratorUser(); err != nil {
		return err
	}

	logger.Info("store latest configuration...\n")
	if err := storeLatestConfiguration(); err != nil {
		return err
	}

	logger.Info("generate default shortcut section...\n")
	if err := generateDefaultShortcutSection(); err != nil {
		return err
	}

	logger.Info("fetch user agent list...\n")
	if err := fetchUserAgent(); err != nil {
		return err
	}

	return nil
}

func Start(ctx context.Context, listener *net.Listener) {
	monitor_realtime.Loop(ctx, time.Second)
	monitor_process_realtime.Loop(ctx, time.Second)
	user_notification.StartListenUserNotificationNotify(ctx)

	go func() {
		if err := startServer(listener, configuration.Get().ServerMonitor.Development.Enable); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				logger.Info("normal shutdown\n")
			} else {
				logger.Fatal("start server failed, %w\n", errors.New(err))
			}
		}
	}()
}

func Stop(ctx context.Context) {
	if err := stopServer(ctx); err != nil {
		logger.Warn("stop server failed, but whatever, %w\n", errors.New(err))
	}

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
			return errors.Errorf("copy failed, %w", err)
		}

		db.Where(map[string]interface{}{"CPU": cpuInfo.CPU}).Attrs(cpuInfo).FirstOrCreate(&cpuInfo)
	}

	for _, v := range systemStat.Disk {
		diskInfo := monitor_model.StoredSystemDiskInfo{}
		if err := copier.Copy(&diskInfo, &v.PartitionStat); err != nil {
			return errors.Errorf("copy failed, %w", err)
		}

		db.Where(map[string]interface{}{"Mountpoint": diskInfo.Mountpoint}).Attrs(diskInfo).FirstOrCreate(&diskInfo)
	}

	return nil
}

// 从配置文件中的管理员配置中生成管理员账号.
// 如果数据库中已存在管理员账号, 则检查账号密码是否一致, 不一致则删除账号并重新创建管理员账号.
func generateAdministratorUser() error {
	administrator := configuration.Get().ServerMonitor.Administrator

	if len(administrator.Password) <= 0 || len(administrator.Username) <= 0 {
		return errors.Errorf("administrator username or password is empty. please check config file")
	}

	user, err := monitor_service.GetUser(monitor_model.User{Role: monitor_model.RoleAdministrator})
	if err != nil && !errors.Is(err, monitor_service.ErrorNotFound) {
		return err
	}

	if user.Username == administrator.Username && user.Password == administrator.Password {
		return nil
	}

	if !errors.Is(err, monitor_service.ErrorNotFound) {
		// 管理员账号密码不一致则删除账号并重新创建管理员账号
		if err := monitor_service.DeleteUser(user); err != nil {
			return errors.New(err)
		}
	}

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

	return nil
}

// 存储最新的配置信息到数据库中
func storeLatestConfiguration() error {
	currentConfig := configuration.Get()

	config, err := monitor_service.LatestConfiguration()
	if err != nil {
		return err
	}

	// ModificationTime 相同时认为配置文件未更新
	if config.Configuration.ModificationTime == currentConfig.ModificationTime {
		return nil
	}

	if err := monitor_service.CreateConfiguration(monitor_model.StoredConfiguration{Configuration: *configuration.Get()}); err != nil {
		return err
	}

	return nil
}

// 创建默认的 monitor_model.ShortcutSection 数据
func generateDefaultShortcutSection() error {
	count, err := monitor_service.CountShortcutSection(monitor_model.ShortcutSection{})
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	_, err = monitor_service.CreateOrUpdateShortcutSections([]monitor_model.ShortcutSection{
		{Name: "Default Folder", Default: true},
	})
	if err != nil {
		return err
	}

	return nil
}

// 第一次启动时拉取 userAgent 列表, 并存储到数据库中
func fetchUserAgent() error {
	count, err := monitor_service.CountUserAgent(monitor_model.UserAgent{})
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	if err := monitor_service.RefreshUserAgent(); err != nil {
		return err
	}

	return nil
}
