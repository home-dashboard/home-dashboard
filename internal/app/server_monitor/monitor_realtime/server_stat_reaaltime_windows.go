package monitor_realtime

import (
	"golang.org/x/sys/windows"
	"strings"
	"unsafe"
)

func getAdaptersInfo() (SystemNetworkAdapterInfoList, error) {
	var nai []SystemNetworkAdapterInfo
	var ol uint32
	err := windows.GetAdaptersInfo(nil, &ol)
	if err == nil || ol == 0 {
		return nai, err
	}
	buf := make([]byte, int(ol))
	ai := (*windows.IpAdapterInfo)(unsafe.Pointer(&buf[0]))
	if err := windows.GetAdaptersInfo(ai, &ol); err != nil {
		return nai, err
	}

	for ; ai != nil; ai = ai.Next {
		nai = append(nai, SystemNetworkAdapterInfo{
			Index:       ai.Index,
			Description: strings.Trim(string(ai.Description[:]), " \t\n\000"),
			Type:        int(ai.Type),
		})
	}

	return nai, err
}
