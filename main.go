package main

import (
	"context"
	"log"
	"net"
	"time"
)

const _host = "google.com"

func main() {
	p, cleanup, err := InitPinger()
	if err != nil {
		log.Fatalf("failed to init pinger: %s", err)
	}
	defer cleanup()

	ctx := context.Background()
	addrs, err := resolveHostWithTimeout(ctx, _host)
	if err != nil {
		log.Fatalf("failed to resolve host %s: %s", _host, err)
	}

	for i := 0; i < len(addrs); {
		if addrs[i].IP.To4() == nil {
			// Pinger is initialized with IPv4 addr. Pinger has either IPv4 or IPv6 connection.
			addrs = append(addrs[:i], addrs[i+1:]...)
		} else {
			i++
		}
	}

	if len(addrs) == 0 {
		log.Fatalf("host %s hasn't IPv4 addresses", _host)
	}
	log.Printf("pinging %s with addrs %v", _host, addrs)

	for _, addr := range addrs {
		if err = p.Run(ctx, addr, 1*time.Second, 3); err != nil {
			log.Fatalf("failed to ping host %s with addr %s: %s", _host, addr.String(), err)
		}
	}
}

func resolveHostWithTimeout(ctx context.Context, host string) ([]net.IPAddr, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return net.DefaultResolver.LookupIPAddr(ctx, host)
}
