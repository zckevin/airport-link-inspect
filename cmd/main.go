package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/Dreamacro/clash/constant"
	clashLog "github.com/Dreamacro/clash/log"
	"github.com/samber/lo"
	airportlinkinspect "github.com/zckevin/airport-link-inspect"
	"github.com/zckevin/airport-link-inspect/internal"
)

func init() {
	clashLog.SetLevel(clashLog.WARNING)
}

func getProxyByMatcher(configFile, matcher string) []constant.Proxy {
	links, err := airportlinkinspect.GetLinkConfigMatchesRegex(configFile, matcher)
	if err != nil {
		panic(err)
	}
	// return links[0].Proxy
	return lo.Map(links, func(link *airportlinkinspect.LinkConfig, _ int) constant.Proxy {
		return link.Proxy
	})
}

func main() {
	configFile := flag.String("config", "config.yaml", "the path to the clash config file")
	matcher := flag.String("matcher", "", "the regex to match the proxy name")
	isDebug := flag.Bool("debug", false, "enable pprof server")
	flag.Parse()

	if *isDebug {
		internal.SpawnPprofServer()
	}

	links, err := airportlinkinspect.GetLinkConfigMatchesRegex(*configFile, *matcher)
	if err != nil {
		panic(err)
	}

	results := internal.MapAllConcurrentlyAllSettled(context.Background(), links,
		func(ctx context.Context, link *airportlinkinspect.LinkConfig) (*airportlinkinspect.LinkInfo, error) {
			return airportlinkinspect.ScrapeLinkInfo(link)
		})
	output, err := json.Marshal(internal.UnwrapAll(results))
	if err != nil {
		panic(err)
	}
	fmt.Println(string(output))
}
