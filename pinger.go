package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/icmp"
)

type (
	ICMPPinger interface {
		Ping(ctx context.Context, ipAddr net.IPAddr) (*icmp.Message, error)
	}

	Printer interface {
		Printf(format string, args ...interface{})
	}
)

type Pinger struct {
	icmpPinger ICMPPinger
	printer    Printer
}

func NewPinger(icmpPinger ICMPPinger, printer Printer) *Pinger {
	return &Pinger{
		icmpPinger: icmpPinger,
		printer:    printer,
	}
}

func (p *Pinger) Run(ctx context.Context, addr net.IPAddr, sleep time.Duration, count int) error {
	// Stub to prevent small DDOS-attack.
	if sleep < 1*time.Second {
		return fmt.Errorf("sleep is less than 1 second: %s", sleep)
	}

	i := 0
	for {
		msg, err := p.icmpPinger.Ping(ctx, addr)
		if err != nil {
			return fmt.Errorf("failed to ping %s: %w", addr.String(), err)
		}
		p.printer.Printf("[%s] successfully pinged: %+v %+v", addr.String(), msg, msg.Body)

		if count > 0 {
			i++
			if i == count {
				return nil
			}
		}

		if err := sleepWithCtx(ctx, sleep); err != nil {
			return err
		}
	}
}

func sleepWithCtx(ctx context.Context, sleep time.Duration) error {
	t := time.NewTimer(sleep)
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		t.Stop()
		return ctx.Err()
	}
}
