package main

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Environment string `yaml:"environment"`
	Host        string `yaml:"host"`
	Port        uint16 `yaml:"port"`
	Redis       struct {
		Host                 string        `yaml:"host"`
		Port                 uint16        `yaml:"port"`
		Username             string        `yaml:"username"`
		Password             string        `yaml:"password"`
		Database             int           `yaml:"database"`
		JavaCacheDuration    time.Duration `yaml:"java_cache_duration"`
		BedrockCacheDuration time.Duration `yaml:"bedrock_cache_duration"`
		IconCacheDuration    time.Duration `yaml:"icon_cache_duration"`
	} `yaml:"redis"`
}

func (c *Config) ReadFile(file string) error {
	data, err := os.ReadFile(file)

	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, c)
}
