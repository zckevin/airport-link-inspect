package tcplinkinspect

import (
	"fmt"
	"os"
	"regexp"

	"github.com/Dreamacro/clash/config"
	"github.com/Dreamacro/clash/constant"
)

type ProxyRawInfo map[string]any

type LinkConfig struct {
	Desc    string
	RawInfo ProxyRawInfo
	Proxy   constant.Proxy
}

func newLinkConfig(config *ConfigFile, name string) (*LinkConfig, error) {
	rawInfo := config.getRawInfo(name)
	proxy := config.getConfig(name)
	if rawInfo == nil || proxy == nil {
		return nil, fmt.Errorf("no proxy found with name: %s", name)
	}
	return &LinkConfig{RawInfo: rawInfo, Proxy: proxy, Desc: name}, nil
}

type ConfigFile struct {
	RawConfig *config.RawConfig
	Config    *config.Config
}

func (config *ConfigFile) getMatchedNames(regex *regexp.Regexp) []string {
	var names []string
	for _, mapping := range config.RawConfig.Proxy {
		name := mapping["name"].(string)
		if regex.MatchString(name) {
			names = append(names, name)
		}
	}
	return names
}

func (config *ConfigFile) getRawInfo(targetName string) ProxyRawInfo {
	for _, mapping := range config.RawConfig.Proxy {
		if mapping["name"] == targetName {
			return mapping
		}
	}
	return nil
}

func (config *ConfigFile) getConfig(targetName string) constant.Proxy {
	return config.Config.Proxies[targetName]
}

func readConfig(configPath string) (*ConfigFile, error) {
	buf, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	rawConfig, err := config.UnmarshalRawConfig(buf)
	if err != nil {
		return nil, err
	}
	config, err := config.Parse(buf)
	if err != nil {
		return nil, err
	}
	return &ConfigFile{RawConfig: rawConfig, Config: config}, nil
}

func GetLinkConfigWithName(configPath string, name string) (*LinkConfig, error) {
	config, err := readConfig(configPath)
	if err != nil {
		return nil, err
	}
	return newLinkConfig(config, name)
}

func GetLinkConfigMatchesRegex(configPath string, regexString string) ([]*LinkConfig, error) {
	config, err := readConfig(configPath)
	if err != nil {
		return nil, err
	}
	regex, err := regexp.Compile(regexString)
	if err != nil {
		return nil, err
	}
	names := config.getMatchedNames(regex)
	configs := make([]*LinkConfig, len(names))
	for i, name := range names {
		config, err := newLinkConfig(config, name)
		if err != nil {
			return nil, err
		}
		configs[i] = config
	}
	return configs, nil
}
