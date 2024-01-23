//go:build !windows

package monitor_realtime

import "github.com/go-errors/errors"

func getAdaptersInfo() (SystemNetworkAdapterInfoList, error) {
	return nil, errors.New("not implement!")
}
