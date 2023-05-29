package scrapers

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/Dreamacro/clash/constant"
	"github.com/zckevin/tcp-link-inspect/internal"
)

const (
	IPSB_META_INFO_URL = "https://api.ip.sb/geoip/%s"
)

// "organization":"WeiYi Network Technology Co.",
// "longitude":113.722,
// "timezone":"Asia\/Shanghai",
// "isp":"China Telecom Guangdong",
// "offset":28800,
// "asn":63695,
// "asn_organization":"WeiYi Network Technology Co., Ltd",
// "country":"China",
// "ip":"211.99.96.14",
// "latitude":34.7732,
// "continent_code":"AS",
type IpsbMetaInfo struct {
	Organization    string  `json:"organization"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	Timezone        string  `json:"timezone"`
	Isp             string  `json:"isp"`
	Offset          int     `json:"offset"`
	Asn             int     `json:"asn"`
	AsnOrganization string  `json:"asn_organization"`
	Country         string  `json:"country"`
	IP              string  `json:"ip"`
	ContinentCode   string  `json:"continent_code"`
}

func (info *IpsbMetaInfo) ToJSON() []byte {
	b, _ := json.Marshal(info)
	return b
}

func ScrapeMetainfoFromIpsb(ctx context.Context, p constant.Proxy, ip net.IP) (*IpsbMetaInfo, error) {
	u := fmt.Sprintf(IPSB_META_INFO_URL, ip.String())
	body, err := internal.FetchWithProxy(ctx, p, u)
	if err != nil {
		return nil, err
	}
	var meta IpsbMetaInfo
	if err := json.Unmarshal(body, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}
