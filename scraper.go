package tcplinkinspect

import (
	"context"
	"fmt"
	"net"

	"github.com/zckevin/tcp-link-inspect/internal"
	"github.com/zckevin/tcp-link-inspect/scrapers"
)

func parseAccessPoint(resolverName string, link *LinkConfig) ([]IPMetaInfo, error) {
	server, ok := link.RawInfo["server"].(string)
	if !ok {
		return nil, fmt.Errorf("no server found in yaml config")
	}
	ips, err := net.LookupIP(server)
	if err != nil {
		return nil, err
	}
	callback := func(ctx context.Context, ip net.IP) (IPMetaInfo, error) {
		switch resolverName {
		case "ipinfo":
			return scrapers.ScrapeMetainfoFromIpinfo(ctx, link.Proxy, ip)
		case "ipsb":
			return scrapers.ScrapeMetainfoFromIpsb(ctx, link.Proxy, ip)
		default:
			panic("unknown resolver name")
		}
	}
	return internal.MapAll(context.Background(), ips, callback)
}

func parseLandingPoint(link *LinkConfig) (IPMetaInfo, error) {
	return scrapers.ScrapeMetainfoFromCloudflare(context.Background(), link.Proxy)
}

func ScrapeLinkInfo(link *LinkConfig) (*LinkInfo, error) {
	result := &LinkInfo{
		Desc: link.Desc,
	}
	accessPointFn := func(ctx context.Context) error {
		infos, err := parseAccessPoint("ipinfo", link)
		if err != nil {
			return err
		}
		// fmt.Printf("%+v\n", infos[0])
		result.AccessPoint = infos
		return nil
	}
	landingPointFn := func(ctx context.Context) error {
		lp, err := parseLandingPoint(link)
		if err != nil {
			return err
		}
		// fmt.Printf("%+v\n", lp)
		result.LandingPoint = lp
		return nil
	}
	rttFn := func(ctx context.Context) error {
		minRtt, err := internal.ProbeLinkMinRtt(1, link.Proxy)
		if err != nil {
			return err
		}
		// fmt.Printf("%+v\n", minRtt)
		result.MinRtt = minRtt
		return nil
	}
	err := internal.PromiseAll(context.Background(), []func(ctx context.Context) error{
		accessPointFn,
		landingPointFn,
		rttFn,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
