package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"runtime"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

type ICMP struct {
	conn         *icmp.PacketConn
	isPrivileged bool
	isV4         bool
}

func NewICMP(network, address string) (*ICMP, error) {
	conn, err := icmp.ListenPacket(network, address)
	if err != nil {
		return nil, err
	}

	return &ICMP{
		conn:         conn,
		isPrivileged: network != "udp4" && network != "udp6",
		isV4:         conn.IPv4PacketConn() != nil,
	}, nil
}

func (i *ICMP) Close() error {
	return i.conn.Close()
}

func (i *ICMP) Ping(ctx context.Context, ipAddr net.IPAddr) (*icmp.Message, error) {
	wm := icmp.Message{
		Type: getICMPTypeEcho(i.isV4),
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  0,
			Data: []byte("HELLO-R-U-THERE"),
		},
	}
	wb, err := wm.Marshal(nil)
	if err != nil {
		return nil, fmt.Errorf("marshal ping packet: %w", err)
	}

	deadline, _ := ctx.Deadline()
	if err = i.conn.SetDeadline(deadline); err != nil {
		return nil, fmt.Errorf("set deadline %s: %w", deadline, err)
	}
	addr := net.Addr(&ipAddr)
	if !i.isPrivileged {
		addr = net.Addr(&net.UDPAddr{IP: ipAddr.IP})
	}
	if _, err = i.conn.WriteTo(wb, addr); err != nil {
		return nil, fmt.Errorf("write ping packet to %s: %w", addr, err)
	}

	rb := make([]byte, 1500)
	n, peer, err := i.conn.ReadFrom(rb)
	if err != nil {
		return nil, fmt.Errorf("read from conn: %w", err)
	}
	if !equalAddresses(peer, addr) {
		// Someone requested us. Usually it cannot happen.
		return nil, fmt.Errorf("got reply from not requested addr %s, requested %s", peer, addr)
	}

	// XXX(ArtAndreev): ReadFrom is bugged on Darwin, it returns part of IP header if ping isn't privileged.
	//  See https://github.com/golang/go/issues/47369.
	offset := 0
	if (runtime.GOOS == "darwin" || runtime.GOOS == "ios") && !i.isPrivileged {
		offset = 20
	}
	rm, err := icmp.ParseMessage(getICMPProtoNumber(i.isV4), rb[offset:n])
	if err != nil {
		return nil, fmt.Errorf("parse message: %w", err)
	}
	if i.isV4 {
		if rm.Type == ipv4.ICMPTypeEchoReply {
			return rm, nil
		}
	} else {
		if rm.Type == ipv6.ICMPTypeEchoReply {
			return rm, nil
		}
	}
	return nil, fmt.Errorf("got %+v, want echo reply", rm)
}
