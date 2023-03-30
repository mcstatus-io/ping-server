package main

import (
	"errors"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration.
type Config struct {
	Environment string  `yaml:"environment"`
	Host        string  `yaml:"host"`
	Port        uint16  `yaml:"port"`
	Redis       *string `yaml:"redis"`
	Cache       struct {
		JavaStatusDuration    time.Duration `yaml:"java_status_duration" json:"java_status_duration"`
		BedrockStatusDuration time.Duration `yaml:"bedrock_status_duration" json:"bedrock_status_duration"`
		IconDuration          time.Duration `yaml:"icon_duration" json:"icon_duration"`
	} `yaml:"cache"`
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
