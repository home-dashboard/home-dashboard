//go:build !windows

package monitor_realtime

import (
	"errors"
)

func getAdaptersInfo() (SystemNetworkAdapterInfoList, error) {
	return nil, errors.New("not implement!")
}
