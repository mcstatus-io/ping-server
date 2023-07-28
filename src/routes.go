package main

import (
	"encoding/json"
	"fmt"
	"main/src/assets"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/mcstatus-io/mcutil"
)

func init() {
	app.Get("/ping", PingHandler)
	app.Get("/favicon.ico", FaviconHandler)
	app.Get("/status/java/:address", JavaStatusHandler)
	app.Get("/status/bedrock/:address", BedrockStatusHandler)
	app.Get("/icon", DefaultIconHandler)
	app.Get("/icon/:address", IconHandler)
	app.Get("/debug/java/:address", DebugJavaStatusHandler)
	app.Get("/debug/legacy/:address", DebugLegacyStatusHandler)
	app.Get("/debug/bedrock/:address", DebugBedrockStatusHandler)
	app.Get("/debug/query/basic/:address", DebugBasicQueryHandler)
	app.Get("/debug/query/full/:address", DebugFullQueryHandler)
	app.Use(NotFoundHandler)
}

// PingHandler responds with a 200 OK status for simple health checks.
func PingHandler(ctx *fiber.Ctx) error {
	return ctx.SendStatus(http.StatusOK)
}

// FaviconHandler serves the favicon.ico file to any users that visit the API using a browser.
func FaviconHandler(ctx *fiber.Ctx) error {
	return ctx.Type("ico").Send(assets.Favicon)
}

// JavaStatusHandler returns the status of the Java edition Minecraft server specified in the address parameter.
func JavaStatusHandler(ctx *fiber.Ctx) error {
	host, port, err := ParseAddress(ctx.Params("address"), 25565)

	if err != nil {
		return ctx.Status(http.StatusBadRequest).SendString("Invalid address value")
	}

	if err = r.Increment(fmt.Sprintf("java-hits:%s-%d", host, port)); err != nil {
		return err
	}

	response, expiresAt, err := GetJavaStatus(host, port, ctx.QueryBool("query", true))

	if err != nil {
		return err
	}

	ctx.Set("X-Cache-Hit", strconv.FormatBool(expiresAt != 0))

	if expiresAt != 0 {
		ctx.Set("X-Cache-Time-Remaining", strconv.Itoa(int(expiresAt.Seconds())))
	}

	return ctx.JSON(response)
}

// BedrockStatusHandler returns the status of the Bedrock edition Minecraft server specified in the address parameter.
func BedrockStatusHandler(ctx *fiber.Ctx) error {
	host, port, err := ParseAddress(ctx.Params("address"), 19132)

	if err != nil {
		return ctx.Status(http.StatusBadRequest).SendString("Invalid address value")
	}

	if err = r.Increment(fmt.Sprintf("bedrock-hits:%s-%d", host, port)); err != nil {
		return err
	}

	response, expiresAt, err := GetBedrockStatus(host, port)

	if err != nil {
		return err
	}

	ctx.Set("X-Cache-Hit", strconv.FormatBool(expiresAt != 0))

	if expiresAt != 0 {
		ctx.Set("X-Cache-Time-Remaining", strconv.Itoa(int(expiresAt.Seconds())))
	}

	return ctx.JSON(response)
}

// IconHandler returns the server icon for the specified Java edition Minecraft server.
func IconHandler(ctx *fiber.Ctx) error {
	host, port, err := ParseAddress(ctx.Params("address"), 25565)

	if err != nil {
		return ctx.Status(http.StatusBadRequest).SendString("Invalid address value")
	}

	icon, expiresAt, err := GetServerIcon(host, port)

	if err != nil {
		return err
	}

	ctx.Set("X-Cache-Hit", strconv.FormatBool(expiresAt != 0))

	if expiresAt != 0 {
		ctx.Set("X-Cache-Time-Remaining", strconv.Itoa(int(expiresAt.Seconds())))
	}

	return ctx.Type("png").Send(icon)
}

// DefaultIconHandler returns the default server icon.
func DefaultIconHandler(ctx *fiber.Ctx) error {
	return ctx.Type("png").Send(assets.DefaultIcon)
}

// DebugJavaStatusHandler returns the status of the Java edition Minecraft server specified in the address parameter.
func DebugJavaStatusHandler(ctx *fiber.Ctx) error {
	host, port, err := ParseAddress(ctx.Params("address"), 25565)

	if err != nil {
		return ctx.Status(http.StatusBadRequest).SendString("Invalid address value")
	}

	if err = r.Increment(fmt.Sprintf("java-hits:%s-%d", host, port)); err != nil {
		return err
	}

	result, err := mcutil.StatusRaw(host, port)

	if err != nil {
		return ctx.Status(http.StatusBadRequest).SendString(err.Error())
	}

	data, err := json.MarshalIndent(result, "", "    ")

	if err != nil {
		return err
	}

	return ctx.Type("json").Send(data)
}

// DebugJavaStatusHandler returns the legacy status of the Java edition Minecraft server specified in the address parameter.
func DebugLegacyStatusHandler(ctx *fiber.Ctx) error {
	host, port, err := ParseAddress(ctx.Params("address"), 25565)

	if err != nil {
		return ctx.Status(http.StatusBadRequest).SendString("Invalid address value")
	}

	if err = r.Increment(fmt.Sprintf("java-hits:%s-%d", host, port)); err != nil {
		return err
	}

	result, err := mcutil.StatusLegacy(host, port)

	if err != nil {
		return ctx.Status(http.StatusBadRequest).SendString(err.Error())
	}

	data, err := json.MarshalIndent(result, "", "    ")

	if err != nil {
		return err
	}

	return ctx.Type("json").Send(data)
}

// DebugBedrockStatusHandler returns the status of the Bedrock edition Minecraft server specified in the address parameter.
func DebugBedrockStatusHandler(ctx *fiber.Ctx) error {
	host, port, err := ParseAddress(ctx.Params("address"), 19132)

	if err != nil {
		return ctx.Status(http.StatusBadRequest).SendString("Invalid address value")
	}

	if err = r.Increment(fmt.Sprintf("bedrock-hits:%s-%d", host, port)); err != nil {
		return err
	}

	result, err := mcutil.StatusBedrock(host, port)

	if err != nil {
		return ctx.Status(http.StatusBadRequest).SendString(err.Error())
	}

	data, err := json.MarshalIndent(result, "", "    ")

	if err != nil {
		return err
	}

	return ctx.Type("json").Send(data)
}

// DebugBasicQueryHandler returns the basic query information of the Java edition Minecraft server specified in the address parameter.
func DebugBasicQueryHandler(ctx *fiber.Ctx) error {
	host, port, err := ParseAddress(ctx.Params("address"), 25565)

	if err != nil {
		return ctx.Status(http.StatusBadRequest).SendString("Invalid address value")
	}

	if err = r.Increment(fmt.Sprintf("bedrock-hits:%s-%d", host, port)); err != nil {
		return err
	}

	result, err := mcutil.BasicQuery(host, port)

	if err != nil {
		return ctx.Status(http.StatusBadRequest).SendString(err.Error())
	}

	data, err := json.MarshalIndent(result, "", "    ")

	if err != nil {
		return err
	}

	return ctx.Type("json").Send(data)
}

// DebugFullQueryHandler returns the full query information of the Java edition Minecraft server specified in the address parameter.
func DebugFullQueryHandler(ctx *fiber.Ctx) error {
	host, port, err := ParseAddress(ctx.Params("address"), 25565)

	if err != nil {
		return ctx.Status(http.StatusBadRequest).SendString("Invalid address value")
	}

	if err = r.Increment(fmt.Sprintf("bedrock-hits:%s-%d", host, port)); err != nil {
		return err
	}

	result, err := mcutil.FullQuery(host, port)

	if err != nil {
		return ctx.Status(http.StatusBadRequest).SendString(err.Error())
	}

	data, err := json.MarshalIndent(result, "", "    ")

	if err != nil {
		return err
	}

	return ctx.Type("json").Send(data)
}

// NotFoundHandler handles requests to routes that do not exist and returns a 404 Not Found status.
func NotFoundHandler(ctx *fiber.Ctx) error {
	return ctx.SendStatus(http.StatusNotFound)
}
