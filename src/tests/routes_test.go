package main

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
)

func TestPingHandler(t *testing.T) {
	app := fiber.New()

	app.Get("/ping", PingHandler)

	resp, err := app.Test(utils.NewRequest(http.MethodGet, "/ping"))
	utils.AssertEqual(t, nil, err, "PingHandler")
	utils.AssertEqual(t, http.StatusOK, resp.StatusCode, "PingHandler")
}

func TestJavaStatusHandler(t *testing.T) {
	app := fiber.New()

	app.Get("/status/java/:address", JavaStatusHandler)

	resp, err := app.Test(utils.NewRequest(http.MethodGet, "/status/java/localhost:25565"))
	utils.AssertEqual(t, nil, err, "JavaStatusHandler")
	utils.AssertEqual(t, http.StatusOK, resp.StatusCode, "JavaStatusHandler")

	// Test invalid address
	resp, err = app.Test(utils.NewRequest(http.MethodGet, "/status/java/invalid"))
	utils.AssertEqual(t, nil, err, "JavaStatusHandler")
	utils.AssertEqual(t, http.StatusBadRequest, resp.StatusCode, "JavaStatusHandler")
}

func TestBedrockStatusHandler(t *testing.T) {
	app := fiber.New()

	app.Get("/status/bedrock/:address", BedrockStatusHandler)

	resp, err := app.Test(utils.NewRequest(http.MethodGet, "/status/bedrock/localhost:19132"))
	utils.AssertEqual(t, nil, err, "BedrockStatusHandler")
	utils.AssertEqual(t, http.StatusOK, resp.StatusCode, "BedrockStatusHandler")

	// Test invalid address
	resp, err = app.Test(utils.NewRequest(http.MethodGet, "/status/bedrock/invalid"))
	utils.AssertEqual(t, nil, err, "BedrockStatusHandler")
	utils.AssertEqual(t, http.StatusBadRequest, resp.StatusCode, "BedrockStatusHandler")
}

func TestIconHandler(t *testing.T) {
	app := fiber.New()

	app.Get("/icon/:address", IconHandler)

	resp, err := app.Test(utils.NewRequest(http.MethodGet, "/icon/localhost:25565"))
	utils.AssertEqual(t, nil, err, "IconHandler")
	utils.AssertEqual(t, http.StatusOK, resp.StatusCode, "IconHandler")

	// Test invalid address
	resp, err = app.Test(utils.NewRequest(http.MethodGet, "/icon/invalid"))
	utils.AssertEqual(t, nil, err, "IconHandler")
	utils.AssertEqual(t, http.StatusBadRequest, resp.StatusCode, "IconHandler")
}
