package main

import (
	"errors"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// CacheConfig contains cache-related settings.
type CacheConfig struct {
	JavaStatusDuration    time.Duration `yaml:"java_status_duration" json:"java_status_duration"`
	BedrockStatusDuration time.Duration `yaml:"bedrock_status_duration" json:"bedrock_status_duration"`
	IconDuration          time.Duration `yaml:"icon_duration" json:"icon_duration"`
}

// Config represents the application configuration.
type Config struct {
	Environment string      `yaml:"environment"`
	Host        string      `yaml:"host"`
	Port        uint16      `yaml:"port"`
	Redis       *string     `yaml:"redis"`
	Cache       CacheConfig `yaml:"cache"`
}

// ReadFile reads the configuration from the given file and overrides values using environment variables.
func (c *Config) ReadFile(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, c); err != nil {
		return err
	}

	return c.overrideWithEnvVars()
}

// overrideWithEnvVars overrides configuration values using environment variables.
func (c *Config) overrideWithEnvVars() error {
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		c.Environment = env
	}

	if host := os.Getenv("HOST"); host != "" {
		c.Host = host
	}

	if port := os.Getenv("PORT"); port != "" {
		portInt, err := strconv.Atoi(port)
		if err != nil {
			return errors.New("invalid port value in environment variable")
		}
		c.Port = uint16(portInt)
	}
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		c.Redis = &redisURL
	}

	return nil
}
