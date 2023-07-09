# airport-link-inspect

## Description

Inspect link access/landing point metadata for Clash users,
especially for services using IEPL/IPLC forwarding.

It would return the following information:

- desc: name of the proxy
- proxy_info: proxy info in Clash config file
- access_point: domestic access server's ip metadata
- landing_point: overseas landing server's ip metadata
- min_rtt: minimum round trip time(ms) probing https://www.gstatic.com/generate_204
- supports_udp: if proxy supports udp forwarding

## Usage

```bash
go run ./cmd/main.go -config $HOME/.config/clash/config.yaml -matcher 香港 | jq
```

## Example output

```json
[{
  "desc": "香港1",
  "access_point": [
    {
      "code": 200,
      "ip": "134.175.75.151",
      "msg": "success",
      "result": {
        "Country": "中国",
        "CountryCode": "CN",
        "Region": "华南",
        "RegionCode": "800000",
        "Province": "广东省",
        "ProvinceCode": "440000",
        "City": "广州市",
        "CityCode": "440100",
        "Isp": "电信",
        "IspCode": "100017"
      }
    }
  ],
  "landing_point": {
    "asOrganization": "G-Core Labs SA",
    "asn": 199524,
    "city": "Central",
    "clientIP": "xxx.xxx.xxx.xxx",
    "colo": "HKG",
    "country": "HK",
    "hostname": "speed.cloudflare.com",
    "httpProtocol": "HTTP/1.1",
    "latitude": "22.29080",
    "longitude": "114.15010",
    "region": "Central and Western District"
  },
  "min_rtt": 51,
  "supports_udp": true,
  "proxy_info": {
    "cipher": "aes-256-gcm",
    "name": "香港1",
    "password": "my_strong_password",
    "port": 6666,
    "server": "hk.my-server.com",
    "type": "ss"
  }
}]
```