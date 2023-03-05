package main

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type CacheConfig struct {
	JavaStatusDuration    time.Duration `yaml:"java_status_duration" json:"java_status_duration"`
	BedrockStatusDuration time.Duration `yaml:"bedrock_status_duration" json:"bedrock_status_duration"`
	IconDuration          time.Duration `yaml:"icon_duration" json:"icon_duration"`
}

type Config struct {
	Environment string      `yaml:"environment"`
	Host        string      `yaml:"host"`
	Port        uint16      `yaml:"port"`
	Redis       *string     `yaml:"redis"`
	Cache       CacheConfig `yaml:"cache"`
}

func (c *Config) ReadFile(file string) error {
	data, err := os.ReadFile(file)

	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, c)
}
