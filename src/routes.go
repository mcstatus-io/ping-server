package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func init() {
	app.Get("/ping", PingHandler)
	app.Get("/status/java/:address", JavaStatusHandler)
	app.Get("/status/bedrock/:address", BedrockStatusHandler)
	app.Get("/widget/java/:address", JavaWidgetHandler)
	app.Get("/icon", DefaultIconHandler)
	app.Get("/icon/:address", IconHandler)
	app.Use(NotFoundHandler)
}

// PingHandler responds with a 200 OK status for simple health checks.
func PingHandler(ctx *fiber.Ctx) error {
	return ctx.SendStatus(http.StatusOK)
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

	response, expiresAt, err := GetJavaStatus(host, port)

	if err != nil {
		return err
	}

	ctx.Set("X-Cache-Hit", strconv.FormatBool(expiresAt != 0))

	if expiresAt != 0 {
		ctx.Set("X-Cache-Time-Remaining", strconv.Itoa(int(expiresAt.Seconds())))
	}

	return ctx.JSON(response)
}

// JavaWidgetHandler returns the widget of the Java Edition server.
func JavaWidgetHandler(ctx *fiber.Ctx) error {
	host, port, err := ParseAddress(ctx.Params("address"), 25565)

	if err != nil {
		return ctx.Status(http.StatusBadRequest).SendString("Invalid address value")
	}

	if err = r.Increment(fmt.Sprintf("java-hits:%s-%d", host, port)); err != nil {
		return err
	}

	response, _, err := GetJavaStatus(host, port)

	if err != nil {
		return err
	}

	widget, err := GenerateJavaWidget(response, ctx.QueryBool("dark", true))

	if err != nil {
		return err
	}

	return ctx.Type("png").Send(widget)
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
	return ctx.Type("png").Send(defaultIconBytes)
}

// NotFoundHandler handles requests to routes that do not exist and returns a 404 Not Found status.
func NotFoundHandler(ctx *fiber.Ctx) error {
	return ctx.SendStatus(http.StatusNotFound)
}
