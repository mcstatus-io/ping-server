package main

import (
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Host  string `yaml:"host"`
	Port  uint16 `yaml:"port"`
	Redis string `yaml:"redis"`
	Cache struct {
		Enable               bool          `yaml:"enable"`
		JavaCacheDuration    time.Duration `yaml:"java_cache_duration"`
		BedrockCacheDuration time.Duration `yaml:"bedrock_cache_duration"`
		IconCacheDuration    time.Duration `yaml:"icon_cache_duration"`
	} `yaml:"cache"`
}

func (c *Config) ReadFile(file string) error {
	data, err := ioutil.ReadFile(file)

	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, c)
}
