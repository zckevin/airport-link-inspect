package airportlinkinspect

import "encoding/json"

type IPMetaInfo interface {
	ToJSON() []byte
	GetIPAddr() string
}

type LinkInfo struct {
	Name         string       `json:"name"`
	ProxyInfo    ProxyRawInfo `json:"proxy_info"`
	AccessPoint  []IPMetaInfo `json:"access_point"`
	LandingPoint IPMetaInfo   `json:"landing_point"`
	MinRtt       int          `json:"min_rtt"`
	SupportsUDP  bool         `json:"supports_udp"`
}

func (info *LinkInfo) String() string {
	jsonStr, _ := json.MarshalIndent(info, "", "  ")
	return string(jsonStr)
}
