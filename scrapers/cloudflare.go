package scrapers

import (
	"context"
	"encoding/json"

	"github.com/Dreamacro/clash/constant"
	"github.com/zckevin/tcp-link-inspect/internal"
)

const (
	CLOUDFLARE_META_INFO_SCRAPE_URL = "https://speed.cloudflare.com/meta"
)

type CloudflareMetaInfo struct {
	AsOrgnization string `json:"asOrganization"`
	ASN           int    `json:"asn"`
	City          string `json:"city"`
	ClientIp      string `json:"clientIP"`
	Colo          string `json:"colo"`
	Country       string `json:"country"`
	Hostname      string `json:"hostname"`
	HttpProtocol  string `json:"httpProtocol"`
	Latitude      string `json:"latitude"`
	Longitude     string `json:"longitude"`
	Region        string `json:"region"`
}

func (info *CloudflareMetaInfo) ToJSON() []byte {
	b, _ := json.Marshal(info)
	return b
}

func (info *CloudflareMetaInfo) GetIPAddr() string {
	return info.ClientIp
}

func ScrapeMetainfoFromCloudflare(ctx context.Context, p constant.Proxy) (*CloudflareMetaInfo, error) {
	body, err := internal.FetchWithProxy(ctx, p, CLOUDFLARE_META_INFO_SCRAPE_URL)
	if err != nil {
		return nil, err
	}
	var meta CloudflareMetaInfo
	if err := json.Unmarshal(body, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}
