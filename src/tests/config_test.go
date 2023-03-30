package main

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestReadFile tests the ReadFile function of the Config struct.
func TestReadFile(t *testing.T) {
	// Create a temporary YAML configuration file for testing
	tmpFile, err := ioutil.TempFile("", "config.yml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Define a sample YAML content for testing
	yamlContent := `
environment: production
host: 192.168.1.1
port: 8080
cache:
  java_status_duration: 10m
  bedrock_status_duration: 15m
  icon_duration: 30m
`
	// Write the YAML content to the temporary file
	_, err = tmpFile.WriteString(yamlContent)
	assert.NoError(t, err)

	// Create a Config instance
	config := &Config{}

	// Call the ReadFile function
	err = config.ReadFile(tmpFile.Name())
	assert.NoError(t, err)

	// Verify the values in the Config instance
	assert.Equal(t, "production", config.Environment)
	assert.Equal(t, "192.168.1.1", config.Host)
	assert.Equal(t, uint16(8080), config.Port)
	assert.Equal(t, 10*time.Minute, config.Cache.JavaStatusDuration)
	assert.Equal(t, 15*time.Minute, config.Cache.BedrockStatusDuration)
	assert.Equal(t, 30*time.Minute, config.Cache.IconDuration)
}

// TestOverrideWithEnvVars tests the overrideWithEnvVars function of the Config struct.
func TestOverrideWithEnvVars(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("ENVIRONMENT", "development")
	os.Setenv("HOST", "localhost")
	os.Setenv("PORT", "3000")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("HOST")
		os.Unsetenv("PORT")
	}()

	// Create a Config instance with default values
	config := &Config{
		Environment: "production",
		Host:        "192.168.1.1",
		Port:        8080,
	}

	// Call the overrideWithEnvVars function
	err := config.overrideWithEnvVars()
	assert.NoError(t, err)

	// Verify that the values have been overridden by the environment variables
	assert.Equal(t, "development", config.Environment)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, uint16(3000), config.Port)
}

// TestOverrideWithEnvVarsInvalidPort tests the overrideWithEnvVars function with an invalid port value.
func TestOverrideWithEnvVarsInvalidPort(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("PORT", "invalid")
	defer os.Unsetenv("PORT")

	// Create a Config instance
	config := &Config{}

	// Call the overrideWithEnvVars function
	err := config.overrideWithEnvVars()

	// Verify that an error is returned due to the invalid port value
	assert.Error(t, err)
	assert.Equal(t, "invalid port value in environment variable", err.Error())
}
