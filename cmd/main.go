package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/Dreamacro/clash/constant"
	clashLog "github.com/Dreamacro/clash/log"
	"github.com/samber/lo"
	tcplinkinspect "github.com/zckevin/tcp-link-inspect"
	"github.com/zckevin/tcp-link-inspect/http2ping"
	"github.com/zckevin/tcp-link-inspect/internal"
)

func init() {
	clashLog.SetLevel(clashLog.WARNING)
}

func getProxyByMatcher(configFile, matcher string) []constant.Proxy {
	links, err := tcplinkinspect.GetLinkConfigMatchesRegex(configFile, matcher)
	if err != nil {
		panic(err)
	}
	// return links[0].Proxy
	return lo.Map(links, func(link *tcplinkinspect.LinkConfig, _ int) constant.Proxy {
		return link.Proxy
	})
}

func main() {
	internal.SpawnPprofServer()

	configFile := flag.String("config", "config.yaml", "the path to the clash config file")
	// matcher := flag.String("matcher", "", "the regex to match the proxy name")
	flag.Parse()

	/*
		links, err := tcplinkinspect.GetLinkConfigMatchesRegex(*configFile, *matcher)
		if err != nil {
			panic(err)
		}
	*/
	/*
		pinger := http2ping.NewHTTP2PingerWrapper(http2ping.HTTP2_SERVER, links[0].Proxy)
		for {
			pinger.GetSmoothRtt()
			time.Sleep(1 * time.Second)
		}
	*/

	/*
		links := []constant.Proxy{
			// getProxyByMatcher(*configFile, "BBC"),
			// getProxyByMatcher(*configFile, "HK"),
			getProxyByMatcher(*configFile, "新加坡"),
		}
	*/
	links := getProxyByMatcher(*configFile, "HK")
	g, _ := http2ping.NewHTTP2PingGroup(http2ping.HTTP2_SERVER, links, 0.2)
	// go func() {
	// 	time.Sleep(3 * time.Second)
	// 	log.Println("=========")
	// 	g.Debug()
	// }()
	for {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if best := g.GetMinRttProxy(ctx); best != nil {
				fmt.Println("===", best.Name(), time.Now())
			} else {
				fmt.Println("=== nil")
			}
		}()
		time.Sleep(1 * time.Second)
	}

	/*
		results := internal.MapAllConcurrentlyAllSettled(context.Background(), links[0:1],
			func(ctx context.Context, link *tcplinkinspect.LinkConfig) (*tcplinkinspect.LinkInfo, error) {
				return tcplinkinspect.ScrapeLinkInfo(link)
			})
		output, err := json.Marshal(internal.UnwrapAll(results))
		if err != nil {
			panic(err)
		}
		fmt.Println(string(output))
	*/
}
