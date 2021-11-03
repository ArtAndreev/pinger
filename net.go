package main

import (
	"errors"
	"fmt"
	"net"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

func GetFirstV4ExternalIPAddress() (string, error) {
	ips, err := GetExternalIPs()
	if err != nil {
		return "", fmt.Errorf("get external IPs: %w", err)
	}
	for _, ip := range ips {
		ip = ip.To4()
		if ip != nil {
			return ip.String(), nil
		}
	}
	return "", errors.New("no external IPv4 found")
}

func GetExternalIPs() ([]net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var res []net.IP
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			// XXX(ArtAndreev): some interfaces have ip6 addresses, but they aren't valid. Find out why.
			ip = ip.To4()
			if ip == nil {
				continue
			}
			res = append(res, ip)
		}
	}
	if len(res) == 0 {
		return nil, errors.New("no external IPs found")
	}
	return res, nil
}

func getICMPTypeEcho(isV4 bool) icmp.Type {
	if isV4 {
		return ipv4.ICMPTypeEcho
	}
	return ipv6.ICMPTypeEchoRequest
}

const (
	_protocolICMP     = 1
	_protocolIPv6ICMP = 58
)

func getICMPProtoNumber(isV4 bool) int {
	if isV4 {
		return _protocolICMP
	}
	return _protocolIPv6ICMP
}

func equalAddresses(l, r net.Addr) bool {
	return l.Network() == r.Network() && l.String() == r.String()
}
