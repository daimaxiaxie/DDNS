package dnsutil

import (
	"fmt"
	"net"
)

const (
	Other = 0x00
	IPV4  = 32
	IPV6  = 128
	ALLIP = 0xA0
)

func GetIP(interfaceName string, IPType int) []string {
	ifaces, err := net.Interfaces()
	//addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}

	ip := make([]string, 0, 2)
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil || (iface.Name != interfaceName && interfaceName != "") {
			continue
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				//fmt.Println("IPNet ", iface.Name, " : ", v.IP.String())
				_, bits := v.Mask.Size()
				if !v.IP.IsLoopback() && !v.IP.IsPrivate() && v.IP.IsGlobalUnicast() && (bits&IPType > 0) {
					ip = append(ip, v.IP.String())
				}
			case *net.IPAddr:
				fmt.Println("IPAddr : ", v)
			}
		}
	}

	return ip
}
