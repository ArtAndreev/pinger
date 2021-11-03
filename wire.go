//go:build wireinject
// +build wireinject

package main

import (
	"io"
	"log"
	"os"

	"github.com/google/wire"
)

var (
	_icmpSet = wire.NewSet(
		wire.Value(wireNetwork("udp4")),
		provideFirstV4ExternalIPAddress,
		provideICMPWithCleanup,
	)

	_pingerSet = wire.NewSet(
		NewPinger,
		wire.Bind(new(ICMPPinger), new(*ICMP)),
		_icmpSet,
		wire.Bind(new(Printer), new(*log.Logger)),
		_loggerSet,
	)

	// Instead of constructor for log.Logger: log.New(os.Stdout, "", log.LstdFlags), because I wanted to init directly
	// without constructor. The result will be a direct call to log.New.
	_loggerSet = wire.NewSet(
		wire.InterfaceValue(new(io.Writer), os.Stdout),
		wire.Value(""),
		wire.Value(log.LstdFlags),
		log.New,
	)
)

type (
	wireNetwork string
	wireAddress string
)

// provideICMPWithCleanup is a wrapper that adds cleanup to ICMP and resolves type conflict.
func provideICMPWithCleanup(network wireNetwork, address wireAddress) (*ICMP, func(), error) {
	i, err := NewICMP(string(network), string(address))
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		_ = i.Close()
	}
	return i, cleanup, err
}

func provideFirstV4ExternalIPAddress() (wireAddress, error) {
	ip, err := GetFirstV4ExternalIPAddress()
	return wireAddress(ip), err
}

func InitPinger() (*Pinger, func(), error) {
	wire.Build(_pingerSet)
	return nil, nil, nil
}
