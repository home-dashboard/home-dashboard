package overseer

import (
	"github.com/go-errors/errors"
	"github.com/siaikin/home-dashboard/internal/pkg/overseer/fetcher"
	"net/rpc"
)

// upgradeServiceClient 定义 overseer 服务的客户端.
type upgradeServiceClient struct {
	*rpc.Client
	OverseerInstance *Overseer
}

func (c *upgradeServiceClient) Upgrade(fetcherName string) error {
	var reply string
	err := c.Call(upgradeServiceName+".Upgrade", fetcherName, &reply)
	return errors.New(err)
}

func (c *upgradeServiceClient) Status() (Status, error) {
	var reply Status
	if err := c.Call(upgradeServiceName+".Status", "", &reply); err != nil {
		return Status{}, errors.New(err)
	}

	return reply, nil
}

func (c *upgradeServiceClient) LatestVersionInfo() (*fetcher.AssetInfo, error) {
	var reply = fetcher.AssetInfo{}
	if err := c.Call(upgradeServiceName+".LatestVersionInfo", "", &reply); err != nil {
		return &fetcher.AssetInfo{}, errors.New(err)
	}

	return &reply, nil
}
