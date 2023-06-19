package http2ping

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Dreamacro/clash/constant"
	"github.com/zckevin/tcp-link-inspect/internal"
	"golang.org/x/net/http2"
)

const (
	HTTP2_SERVER = "https://gcore.com"
)

var _ Pinger = (*http2PingerWrapper)(nil)

type pingerStatusCode = uint32

const (
	PINGER_STATUS_DEAD pingerStatusCode = iota
	PINGER_STATUS_PINGING
	PINGER_STATUS_IDLE
)

const (
	rttAlpha      = 0.2
	oneMinusAlpha = 1 - rttAlpha
	rttBeta       = 0.25
	oneMinusBeta  = 1 - rttBeta
)

func updateSRtt(sRtt, rtt uint32) uint32 {
	return uint32(float32(sRtt)*oneMinusAlpha + float32(rtt)*rttAlpha)
}

func updateMeanDeviation(meanDeviation, sRtt, rtt uint32) uint32 {
	return uint32(float32(meanDeviation)*oneMinusBeta + float32(internal.Abs(int32(sRtt)-int32(rtt)))*rttBeta)
}

type pingerSharedStatus struct {
	statusCode    atomic.Uint32
	latestRtt     atomic.Uint32
	sRtt          atomic.Uint32
	meanDeviation atomic.Uint32
}

type http2Pinger struct {
	*pingerSharedStatus
	serverURL *url.URL
	proxy     constant.Proxy

	hasRecordedRtt atomic.Bool
	newSRttCh      chan uint32

	ctx       context.Context
	ctxCancel context.CancelFunc
	closed    atomic.Bool
}

func newHTTP2Pinger(serverURLString string, proxy constant.Proxy, status *pingerSharedStatus) *http2Pinger {
	if _, err := url.Parse(serverURLString); err != nil {
		panic(err)
	}
	u, err := url.Parse(serverURLString)
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	p := &http2Pinger{
		pingerSharedStatus: status,
		serverURL:          u,
		proxy:              proxy,
		newSRttCh:          make(chan uint32),
		ctx:                ctx,
		ctxCancel:          cancel,
	}
	p.statusCode.Store(PINGER_STATUS_DEAD)
	go p.pingLoop()
	return p
}

func (p *http2Pinger) doPing(tlsConn *tls.Conn, http2Conn *http2.ClientConn) (uint32, error) {
	start := time.Now()
	tlsConn.SetDeadline(start.Add(time.Second * 1))
	defer tlsConn.SetDeadline(time.Time{})
	err := http2Conn.Ping(p.ctx)
	if err != nil {
		return 0, fmt.Errorf("http2 ping: %w", err)
	}
	return uint32(time.Since(start).Milliseconds()), nil
}

func (p *http2Pinger) Ping(tlsConn *tls.Conn, http2Conn *http2.ClientConn) error {
	p.statusCode.Store(PINGER_STATUS_PINGING)
	rtt, err := p.doPing(tlsConn, http2Conn)
	if err != nil {
		p.statusCode.Store(PINGER_STATUS_DEAD)
		return err
	}
	sRtt := rtt
	meanDeviation := rtt / 2
	if p.hasRecordedRtt.Load() {
		sRtt = updateSRtt(p.sRtt.Load(), rtt)
		meanDeviation = updateMeanDeviation(p.meanDeviation.Load(), sRtt, rtt)
	} else {
		p.hasRecordedRtt.Store(true)
	}
	// log.Println("rtt:", rtt, "sRtt:", sRtt, "meanDeviation:", meanDeviation)
	p.sRtt.Store(sRtt)
	p.latestRtt.Store(rtt)
	p.meanDeviation.Store(meanDeviation)
	p.statusCode.Store(PINGER_STATUS_IDLE)
	select {
	case p.newSRttCh <- sRtt:
	default:
	}
	return nil
}

func (p *http2Pinger) Dial(ctx context.Context) (*tls.Conn, *http2.ClientConn, error) {
	// log.Println("dialing: ", p.serverURL)
	rawConn, err := internal.DialProxyConn(ctx, p.proxy, p.serverURL.String())
	if err != nil {
		return nil, nil, fmt.Errorf("dial proxy conn: %w", err)
	}
	tlsConn := tls.Client(rawConn, &tls.Config{
		ServerName: p.serverURL.Hostname(),
		NextProtos: []string{"h2"},
	})
	// set deadline for protocol handshake
	tlsConn.SetDeadline(time.Now().Add(time.Second))
	defer tlsConn.SetDeadline(time.Time{})
	tr := http2.Transport{}
	http2Conn, err := tr.NewClientConn(tlsConn)
	if err != nil {
		return nil, nil, fmt.Errorf("new client conn: %w", err)
	}
	return tlsConn, http2Conn, nil
}

func (p *http2Pinger) pingLoop() {
	loopFn := func() (err error) {
		tlsConn, http2Conn, err := p.Dial(context.Background())
		if err != nil {
			p.statusCode.Store(PINGER_STATUS_DEAD)
			return err
		}
		defer http2Conn.Close()
		for {
			err = p.Ping(tlsConn, http2Conn)
			if err != nil {
				return err
			}
			time.Sleep(time.Second * 1)
		}
	}
	for {
		_ = loopFn()
		// log.Println("ping loop err:", err)
		if p.closed.Load() {
			return
		}
		time.Sleep(time.Second * 3)
	}
}

func (p *http2Pinger) GetSmoothRtt() uint32 {
	switch p.statusCode.Load() {
	case PINGER_STATUS_DEAD:
		return 0
	case PINGER_STATUS_PINGING:
		fallthrough
	case PINGER_STATUS_IDLE:
		return p.sRtt.Load()
	}
	panic("unreachable")
}

func (p *http2Pinger) Close() error {
	p.closed.Store(true)
	p.ctxCancel()
	return nil
}

type http2PingerWrapper struct {
	proxy        constant.Proxy
	serverURL    string
	pinger       atomic.Pointer[http2Pinger]
	sharedStatus *pingerSharedStatus

	lastGetSmoothRttTimeMu sync.Mutex
	lastGetSmoothRttTime   time.Time
}

func NewHTTP2PingerWrapper(serverURL string, proxy constant.Proxy) Pinger {
	w := &http2PingerWrapper{
		proxy:        proxy,
		serverURL:    serverURL,
		sharedStatus: &pingerSharedStatus{},
	}
	w.createPinger()
	return w
}

func (pw *http2PingerWrapper) createPinger() {
	// log.Println("====== create pinger...")
	oldp := pw.pinger.Swap(newHTTP2Pinger(pw.serverURL, pw.proxy, pw.sharedStatus))
	if oldp != nil {
		oldp.Close()
	}
}

func (pw *http2PingerWrapper) resetIfWakeupFromSuspend() {
	pw.lastGetSmoothRttTimeMu.Lock()
	defer pw.lastGetSmoothRttTimeMu.Unlock()

	threshold := time.Second * 2
	if !pw.lastGetSmoothRttTime.IsZero() && time.Since(pw.lastGetSmoothRttTime) >= threshold {
		pw.createPinger()
	}
	pw.lastGetSmoothRttTime = time.Now()
}

func (pw *http2PingerWrapper) GetSmoothRtt() uint32 {
	pw.resetIfWakeupFromSuspend()
	return pw.pinger.Load().GetSmoothRtt()
}

func (pw *http2PingerWrapper) RacingNextSmoothRtt(ctx context.Context, resultCh chan<- constant.Proxy) {
	pw.resetIfWakeupFromSuspend()

	var result constant.Proxy
	p := pw.pinger.Load()
	select {
	case <-ctx.Done():
		// caller has canceled the racing
	case <-p.newSRttCh:
		result = p.proxy
	}
	select {
	case resultCh <- result:
	default:
	}
}

func (pw *http2PingerWrapper) GetProxy() constant.Proxy {
	return pw.pinger.Load().proxy
}
