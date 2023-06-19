package http2ping

import (
	"context"

	"github.com/Dreamacro/clash/constant"
)

type Pinger interface {
	GetProxy() constant.Proxy
	GetSmoothRtt() uint32
	RacingNextSmoothRtt(ctx context.Context, resultCh chan<- constant.Proxy)
}

type PingerGroup interface {
	GetMinRttProxy(ctx context.Context) constant.Proxy
	Swap(ctx context.Context, current constant.Proxy) constant.Proxy
}
