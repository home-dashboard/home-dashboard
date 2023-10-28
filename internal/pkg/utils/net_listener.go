package utils

import (
	"net"
	"time"
)

// ListenTCPAddress 监听 TCP 地址并返回 Listener.
func ListenTCPAddress(addr string) (net.Listener, error) {
	return net.Listen("tcp", addr)
}

// ListenTCPAddresses 监听多个 TCP 地址并返回多个 Listener.
func ListenTCPAddresses(addresses ...string) ([]net.Listener, error) {
	listeners := make([]net.Listener, 0, len(addresses))

	for _, address := range addresses {

		if listener, err := ListenTCPAddress(address); err != nil {
			return nil, err
		} else {
			listeners = append(listeners, listener)
		}
	}
	return listeners, nil
}

// IsTCPAddressAvailable 判断 TCP 地址是否可用
func IsTCPAddressAvailable(address string) bool {
	if conn, err := net.DialTimeout("tcp", address, time.Millisecond*100); err != nil {
		return false
	} else {
		_ = conn.Close()
		return true
	}
}

// IsTCPAddressAllAvailable 判断传入的 TCP 地址是否都可用. 如果有一个不可用则返回 false.
func IsTCPAddressAllAvailable(addresses ...string) bool {
	for _, address := range addresses {
		if !IsTCPAddressAvailable(address) {
			return false
		}
	}
	return true
}
