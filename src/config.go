package main

import (
	"io/ioutil"
	"time"

	"github.com/go-yaml/yaml"
)

type Configuration struct {
	Environment     string        `yaml:"environment"`
	Host            string        `yaml:"host"`
	Port            uint16        `yaml:"port"`
	Redis           string        `yaml:"redis"`
	CacheEnable     bool          `yaml:"cache_enable"`
	StatusCacheTTL  time.Duration `yaml:"status_cache_ttl"`
	FaviconCacheTTL time.Duration `yaml:"favicon_cache_ttl"`
}

func (c *Configuration) ReadFile(path string) error {
	data, err := ioutil.ReadFile(path)

	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, c)
}
