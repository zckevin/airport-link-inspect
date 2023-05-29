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
	IPINFO_META_INFO_URL = "https://ipinfo.io/%s"
)

// "ip": "91.199.84.177",
// "city": "Hong Kong",
// "region": "Central and Western",
// "country": "HK",
// "loc": "22.2783,114.1747",
// "org": "AS199524 G-Core Labs S.A.",
// "timezone": "Asia/Hong_Kong",
// "readme": "https://ipinfo.io/missingauth"
type IpinfoMetaInfo struct {
	Ip       string `json:"ip"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
	Timezone string `json:"timezone"`
}

func (info *IpinfoMetaInfo) ToJSON() []byte {
	b, _ := json.Marshal(info)
	return b
}

func ScrapeMetainfoFromIpinfo(ctx context.Context, p constant.Proxy, ip net.IP) (*IpinfoMetaInfo, error) {
	u := fmt.Sprintf(IPINFO_META_INFO_URL, ip.String())
	body, err := internal.FetchWithProxy(ctx, p, u)
	if err != nil {
		return nil, err
	}
	var meta IpinfoMetaInfo
	if err := json.Unmarshal(body, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}
