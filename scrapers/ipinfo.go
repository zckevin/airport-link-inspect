package scrapers

import (
	"encoding/json"
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

func (info *IpinfoMetaInfo) GetIPAddr() string {
	return info.Ip
}
