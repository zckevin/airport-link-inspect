package http2ping

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/Dreamacro/clash/constant"
	"github.com/samber/lo"
)

var _ PingerGroup = (*http2PingGroup)(nil)

type http2PingGroup struct {
	pingers   []Pinger
	dieCh     chan struct{}
	tolerance float32

	best         atomic.Value
	hasBest      atomic.Bool
	lastBestTime atomic.Value
}

func NewHTTP2PingGroup(serverURL string, proxies []constant.Proxy, tolerance float32) (PingerGroup, error) {
	pingers := lo.Map(proxies, func(proxy constant.Proxy, _ int) Pinger {
		return NewHTTP2PingerWrapper(serverURL, proxy)
	})
	if len(pingers) == 0 {
		return nil, fmt.Errorf("no pingers")
	}
	if tolerance < 0 {
		return nil, fmt.Errorf("tolerance must not be negative")
	}
	g := &http2PingGroup{
		pingers:   pingers,
		tolerance: tolerance,
		dieCh:     make(chan struct{}),
	}
	go g.loop(time.Second)
	return g, nil
}

func (g *http2PingGroup) loop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-g.dieCh:
			return
		case <-ticker.C:
			var best Pinger
			minRtt := uint32(1<<31 - 1)
			for _, p := range g.pingers {
				if rtt := p.GetSmoothRtt(); rtt > 0 && rtt < minRtt {
					minRtt = rtt
					best = p
				}
			}
			g.lastBestTime.Store(time.Now())
			g.hasBest.Store(best != nil)
			if best != nil {
				g.best.Store(best)
			}
		}
	}
}

func (g *http2PingGroup) resetIfWakeupFromSuspend() {
	if lastBestTime, ok := g.lastBestTime.Load().(time.Time); ok && time.Since(lastBestTime) >= time.Second*2 {
		g.hasBest.Store(false)
	}
}

func (g *http2PingGroup) GetMinRttProxy(ctx context.Context) constant.Proxy {
	g.resetIfWakeupFromSuspend()

	if g.hasBest.Load() {
		return g.best.Load().(Pinger).GetProxy()
	}
	if len(g.pingers) == 0 {
		return nil
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	resultCh := make(chan constant.Proxy, 1)
	for _, pinger := range g.pingers {
		go pinger.RacingNextSmoothRtt(ctx, resultCh)
	}
	return <-resultCh
}

func (g *http2PingGroup) Close() error {
	close(g.dieCh)
	return nil
}
