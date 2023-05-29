package tcplinkinspect

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/pkg/errors"
	"github.com/zckevin/tcp-link-inspect/internal"
	"github.com/zckevin/tcp-link-inspect/scrapers"
)

type ipslice []net.IP

func (ips ipslice) ToString(port int) string {
	var key string
	for _, ip := range ips {
		key += fmt.Sprintf("%s:%d,", ip.String(), port)
	}
	return key
}

func (ips ipslice) isBlocked() bool {
	for _, ip := range ips {
		if ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return true
		}
	}
	return false
}

var (
	mu sync.Mutex
	// use ip as key instead of host, because multiple hosts may share the same ip
	// map key is "ip_1:port,ip_2:port,..."
	history map[string]struct{} = make(map[string]struct{})
)

func upsetHistory(ips ipslice, port int) bool {
	mu.Lock()
	defer mu.Unlock()

	key := ips.ToString(port)
	if _, ok := history[key]; ok {
		return false
	}
	history[key] = struct{}{}
	return true
}

func parseAccessPoint(resolverName string, link *LinkConfig, ips []net.IP) ([]IPMetaInfo, error) {
	callback := func(ctx context.Context, ip net.IP) (IPMetaInfo, error) {
		switch resolverName {
		case "ipinfo":
			return scrapers.ScrapeMetainfoFromIpinfo(ctx, link.Proxy, ip)
		case "ipsb":
			return scrapers.ScrapeMetainfoFromIpsb(ctx, link.Proxy, ip)
		case "ipqaros":
			return scrapers.ScrapeMetainfoFromIpqaros(ctx, link.Proxy, ip)
		default:
			panic("unknown resolver name")
		}
	}
	return internal.MapAllConcurrently(context.Background(), ips, callback)
}

func parseLandingPoint(link *LinkConfig) (IPMetaInfo, error) {
	return scrapers.ScrapeMetainfoFromCloudflare(context.Background(), link.Proxy)
}

func ScrapeLinkInfo(link *LinkConfig) (_ *LinkInfo, err error) {
	logErrorFn := func(err error) {
		internal.Logger.Err(err).
			Str("proxy", link.Proxy.Addr()).
			Str("desc", link.Desc).
			Msg("")
	}
	defer func() {
		if err != nil {
			logErrorFn(err)
		}
	}()

	server, ok := link.RawInfo["server"].(string)
	if !ok {
		return nil, fmt.Errorf("no server found in yaml config")
	}
	ips, err := net.LookupIP(server)
	if err != nil {
		return nil, err
	}
	// if local ip, skip
	if ipslice(ips).isBlocked() {
		return nil, fmt.Errorf("non-public ip")
	}
	// if access point ip(s) are scraped before, skip the scraping process
	if appended := upsetHistory(ips, link.Port()); !appended {
		return nil, nil
	}

	result := &LinkInfo{
		Desc:      link.Desc,
		ProxyInfo: link.RawInfo,
	}
	accessPointFn := func(ctx context.Context) error {
		infos, err := parseAccessPoint("ipqaros", link, ips)
		if err != nil {
			return errors.Wrap(err, "parse access point failed")
		}
		result.AccessPoint = infos
		return nil
	}
	landingPointFn := func(ctx context.Context) error {
		lp, err := parseLandingPoint(link)
		if err != nil {
			return errors.Wrap(err, "parse landing point failed")
		}
		result.LandingPoint = lp
		return nil
	}
	rttFn := func(ctx context.Context) error {
		minRtt, err := internal.ProbeLinkMinRtt(3, link.Proxy)
		if err != nil {
			return errors.Wrap(err, "probe link min rtt failed")
		}
		result.MinRtt = minRtt
		return nil
	}
	err = internal.PromiseAll(context.Background(), []func(ctx context.Context) error{
		accessPointFn,
		landingPointFn,
		rttFn,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
