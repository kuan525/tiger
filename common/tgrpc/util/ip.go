package util

import (
	"net"
)

const (
	localhost = "127.0.0.1"
)

// 获取运行这段代码机器的外部IP，如果无法获取则返回localhost
func ExternaIP() string {
	ifaces, err := net.Interfaces() // 遍历机器上所有的网络接口，例如eth0、wlan0等
	if err != nil {
		return localhost
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 { // 接口是否启动
			continue // interfece down
		}
		if iface.Flags&net.FlagLoopback != 0 { // 是否是回环接口
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return localhost
		}
		for _, addr := range addrs {
			ip := getIpFromAddr(addr)
			if ip == nil {
				continue
			}
			return ip.String()
		}
	}
	return localhost
}

func getIpFromAddr(addr net.Addr) net.IP {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	if ip == nil || ip.IsLoopback() { // 是否是回路地址
		return nil
	}
	ip = ip.To4()
	if ip == nil {
		return nil // not an ipv4 addres
	}
	return ip
}
