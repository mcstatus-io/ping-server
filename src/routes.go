package main

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func init() {
	app.Get("/ping", PingHandler)
	app.Get("/status/java/:address", JavaStatusHandler)
	app.Get("/status/bedrock/:address", BedrockStatusHandler)
	app.Get("/icon/:address", IconHandler)
}

func PingHandler(ctx *fiber.Ctx) error {
	return ctx.SendStatus(http.StatusOK)
}

func JavaStatusHandler(ctx *fiber.Ctx) error {
	host, port, err := ParseAddress(ctx.Params("address"), 25565)

	if err != nil {
		return ctx.Status(http.StatusBadRequest).SendString("Invalid address value")
	}

	response, expiresAt, err := GetJavaStatus(host, port)

	if err != nil {
		return err
	}

	if expiresAt != nil {
		ctx.Set("X-Cache-Time-Remaining", strconv.Itoa(int(expiresAt.Seconds())))
	}

	return ctx.Type("json").SendString(response)
}

func BedrockStatusHandler(ctx *fiber.Ctx) error {
	host, port, err := ParseAddress(ctx.Params("address"), 19132)

	if err != nil {
		return ctx.Status(http.StatusBadRequest).SendString("Invalid address value")
	}

	response, expiresAt, err := GetBedrockStatus(host, port)

	if err != nil {
		return err
	}

	if expiresAt != nil {
		ctx.Set("X-Cache-Time-Remaining", strconv.Itoa(int(expiresAt.Seconds())))
	}

	return ctx.Type("json").SendString(response)
}

func IconHandler(ctx *fiber.Ctx) error {
	host, port, err := ParseAddress(ctx.Params("address"), 25565)

	if err != nil {
		return ctx.Status(http.StatusBadRequest).SendString("Invalid address value")
	}

	icon, err := GetServerIcon(host, port)

	if err != nil {
		return err
	}

	return ctx.Type("png").Send(icon)
}
