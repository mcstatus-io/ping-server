package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInit tests the init function in the main package.
func TestInit(t *testing.T) {
	// Create a temporary YAML configuration file for testing
	tmpFile, err := ioutil.TempFile("", "config.yml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Define a sample YAML content for testing
	yamlContent := `
environment: development
host: 127.0.0.1
port: 8080
cache:
  java_status_duration: 10m
  bedrock_status_duration: 15m
  icon_duration: 30m
`
	// Write the YAML content to the temporary file
	_, err = tmpFile.WriteString(yamlContent)
	assert.NoError(t, err)

	// Set environment variables for testing
	os.Setenv("ENVIRONMENT", "development")
	os.Setenv("HOST", "localhost")
	os.Setenv("PORT", "3000")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("HOST")
		os.Unsetenv("PORT")
	}()

	// Call the init function
	init()

	// Verify that the values have been set correctly in the Config instance
	assert.Equal(t, "development", config.Environment)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, uint16(3000), config.Port)
}

