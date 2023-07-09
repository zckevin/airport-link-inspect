package airportlinkinspect

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/Dreamacro/clash/config"
	"github.com/Dreamacro/clash/constant"
)

type ProxyRawInfo map[string]any

type LinkConfig struct {
	// airport brand name
	GroupName string

	// name in config file
	Name string
	// kvs like address, port, etc.
	RawInfo ProxyRawInfo

	Proxy constant.Proxy
}

func newLinkConfig(config *ConfigFile, groupName, name string) (*LinkConfig, error) {
	var (
		rawInfo = config.getRawInfo(name)
		proxy   = config.getConfig(name)
	)
	if rawInfo == nil || proxy == nil {
		return nil, fmt.Errorf("no proxy found with name: %s", name)
	}
	return &LinkConfig{RawInfo: rawInfo, Proxy: proxy, GroupName: groupName, Name: name}, nil
}

func (link *LinkConfig) Port() int {
	if link.RawInfo != nil {
		if port, ok := link.RawInfo["port"].(int); ok {
			return port
		}
	}
	return 0
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
	return newLinkConfig(config, filepath.Base(configPath), name)
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
		config, err := newLinkConfig(config, filepath.Base(configPath), name)
		if err != nil {
			return nil, err
		}
		configs[i] = config
	}
	return configs, nil
}
