package main

import (
	"net"
)

// GetLocalIP returns the outbound local IP address of this device.
func GetLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		// Fallback to searching interfaces
		addrs, err := net.InterfaceAddrs()
		if err == nil {
			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						return ipnet.IP.String()
					}
				}
			}
		}
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// GetMACAddress retrieves the MAC address of the active network interface.
func GetMACAddress() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	// Try to find the interface matching our local IP first
	localIP := GetLocalIP()
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.String() == localIP {
				if iface.HardwareAddr.String() != "" {
					return iface.HardwareAddr.String()
				}
			}
		}
	}

	// Fallback to first non-empty MAC address
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback == 0 && iface.HardwareAddr.String() != "" {
			return iface.HardwareAddr.String()
		}
	}

	return ""
}
