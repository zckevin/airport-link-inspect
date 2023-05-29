package scrapers

import (
	"encoding/json"
)

const (
	IPQAROS_META_INFO_URL = "https://ip.qaros.com/%s"
)

// "code":200,
// "ip":"183.232.117.11",
// "msg":"success",
//
//	"result":{
//	   "Country":"中国",
//	   "CountryCode":"CN",
//	   "Region":"华南",
//	   "RegionCode":"800000",
//	   "Province":"广东省",
//	   "ProvinceCode":"440000",
//	   "City":"河源市",
//	   "CityCode":"441600",
//	   "Isp":"移动",
//	   "IspCode":"100025"
//	}
type IpqarosMetaInfo struct {
	Code   int    `json:"code"`
	Ip     string `json:"ip"`
	Msg    string `json:"msg"`
	Result struct {
		Country      string `json:"Country"`
		CountryCode  string `json:"CountryCode"`
		Region       string `json:"Region"`
		RegionCode   string `json:"RegionCode"`
		Province     string `json:"Province"`
		ProvinceCode string `json:"ProvinceCode"`
		City         string `json:"City"`
		CityCode     string `json:"CityCode"`
		Isp          string `json:"Isp"`
		IspCode      string `json:"IspCode"`
	} `json:"result"`
}

func (info *IpqarosMetaInfo) ToJSON() []byte {
	b, _ := json.Marshal(info)
	return b
}

func (info *IpqarosMetaInfo) GetIPAddr() string {
	return info.Ip
}
