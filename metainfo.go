package tcplinkinspect

import "encoding/json"

type IPMetaInfo interface {
	ToJSON() []byte
}

type LinkInfo struct {
	Desc         string       `json:"desc"`
	AccessPoint  []IPMetaInfo `json:"access_point"`
	LandingPoint IPMetaInfo   `json:"landing_point"`
	MinRtt       int          `json:"min_rtt"`
}

func (info *LinkInfo) String() string {
	jsonStr, _ := json.MarshalIndent(info, "", "  ")
	return string(jsonStr)
}