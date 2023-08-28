package main

import (
	"errors"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	// DefaultConfig is the default configuration values used by the application.
	DefaultConfig *Config = &Config{
		Environment: "production",
		Host:        "127.0.0.1",
		Port:        3001,
		Redis:       nil,
		Cache: ConfigCache{
			EnableLocks:           true,
			JavaStatusDuration:    time.Minute,
			BedrockStatusDuration: time.Minute,
			IconDuration:          time.Minute * 15,
		},
	}
)

// Config represents the application configuration.
type Config struct {
	Environment string      `yaml:"environment"`
	Host        string      `yaml:"host"`
	Port        uint16      `yaml:"port"`
	Redis       *string     `yaml:"redis"`
	Cache       ConfigCache `yaml:"cache"`
}

// ConfigCache represents the caching durations of various responses.
type ConfigCache struct {
	EnableLocks           bool          `yaml:"enable_locks"`
	JavaStatusDuration    time.Duration `yaml:"java_status_duration"`
	BedrockStatusDuration time.Duration `yaml:"bedrock_status_duration"`
	IconDuration          time.Duration `yaml:"icon_duration"`
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

// WriteFile writes the configuration values to a file.
func (c *Config) WriteFile(file string) error {
	data, err := yaml.Marshal(c)

	if err != nil {
		return err
	}

	return os.WriteFile(file, data, 0777)
}

func (c *Config) overrideWithEnvVars() error {
	if value := os.Getenv("ENVIRONMENT"); value != "" {
		c.Environment = value
	}

	if value := os.Getenv("HOST"); value != "" {
		c.Host = value
	}

	if value := os.Getenv("PORT"); value != "" {
		portInt, err := strconv.Atoi(value)

		if err != nil {
			return errors.New("invalid port value in environment variable")
		}

		c.Port = uint16(portInt)
	}

	if value := os.Getenv("REDIS_URL"); value != "" {
		c.Redis = &value
	}

	return nil
}
