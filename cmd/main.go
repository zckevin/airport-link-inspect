package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/Dreamacro/clash/log"
	tcplinkinspect "github.com/zckevin/tcp-link-inspect"
	"github.com/zckevin/tcp-link-inspect/internal"
)

func init() {
	log.SetLevel(log.WARNING)
}

func main() {
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
	fmt.Printf("%+v\n", internal.UnwrapAll(results))
}
