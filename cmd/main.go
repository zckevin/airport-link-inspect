package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/Dreamacro/clash/constant"
	clashLog "github.com/Dreamacro/clash/log"
	"github.com/samber/lo"
	tcplinkinspect "github.com/zckevin/tcp-link-inspect"
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
	matcher := flag.String("matcher", "", "the regex to match the proxy name")
	flag.Parse()

	links, err := tcplinkinspect.GetLinkConfigMatchesRegex(*configFile, *matcher)
	if err != nil {
		panic(err)
	}

	results := internal.MapAllConcurrentlyAllSettled(context.Background(), links,
		func(ctx context.Context, link *tcplinkinspect.LinkConfig) (*tcplinkinspect.LinkInfo, error) {
			return tcplinkinspect.ScrapeLinkInfo(link)
		})
	output, err := json.Marshal(internal.UnwrapAll(results))
	if err != nil {
		panic(err)
	}
	fmt.Println(string(output))
}
