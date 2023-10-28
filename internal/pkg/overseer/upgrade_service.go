package overseer

import "github.com/siaikin/home-dashboard/internal/pkg/overseer/fetcher"

const (
	upgradeServiceName = "upgradeService"
)

type upgradeService struct {
	OverseerInstance *Overseer
}

// Upgrade 会使用名为 fetcherName 的 fetcher.Fetcher 获取最新的二进制文件并更新程序.
func (s *upgradeService) Upgrade(fetcherName string, reply *string) error {
	if err := s.OverseerInstance.Upgrade(fetcherName); err != nil {
		return err
	}

	*reply = "ok"

	return nil
}

// Status 获取当前的运行状态. 状态枚举参考 Status.
func (s *upgradeService) Status(nothing string, reply *Status) error {
	if status, err := s.OverseerInstance.Status(); err != nil {
		return err
	} else {
		*reply = status
	}

	return nil
}

// LatestVersionInfo 获取最新的版本信息. 如果没有找到最新版本, 则返回 nil.
func (s *upgradeService) LatestVersionInfo(nothing string, reply *fetcher.AssetInfo) error {
	if info, err := s.OverseerInstance.LatestVersionInfo(); err != nil {
		return err
	} else {
		*reply = *info
	}

	return nil
}
