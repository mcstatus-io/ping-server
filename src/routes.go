package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mcstatus-io/shared/status"
	"github.com/mcstatus-io/shared/util"
)

func init() {
	app.Get("/ping", PingHandler)
	app.Get("/status/java/:address", JavaStatusHandler)
	app.Get("/status/bedrock/:address", BedrockStatusHandler)
	app.Get("/icon/default", DefaultIconHandler)
	app.Get("/icon/:address", IconHandler)
	app.Use(NotFoundHandler)
}

func PingHandler(ctx *fiber.Ctx) error {
	return ctx.SendStatus(http.StatusOK)
}

func JavaStatusHandler(ctx *fiber.Ctx) error {
	host, port, err := ParseAddress(ctx.Params("address"), 25565)

	if err != nil {
		return ctx.Status(http.StatusBadRequest).SendString("Invalid address value")
	}

	if err = r.Set(fmt.Sprintf("unique:%s-%d", host, port), time.Now(), 0); err != nil {
		return err
	}

	response, expiresAt, err := status.GetJavaStatus(r, host, port, config.Cache.JavaCacheDuration)

	if err != nil {
		return err
	}

	if expiresAt != nil {
		ctx.Set("X-Cache-Time-Remaining", strconv.Itoa(int(expiresAt.Seconds())))
	}

	return ctx.JSON(response)
}

func BedrockStatusHandler(ctx *fiber.Ctx) error {
	host, port, err := ParseAddress(ctx.Params("address"), 19132)

	if err != nil {
		return ctx.Status(http.StatusBadRequest).SendString("Invalid address value")
	}

	if err = r.Set(fmt.Sprintf("unique:%s-%d", host, port), time.Now(), 0); err != nil {
		return err
	}

	response, expiresAt, err := status.GetBedrockStatus(r, host, port, config.Cache.BedrockCacheDuration)

	if err != nil {
		return err
	}

	if expiresAt != nil {
		ctx.Set("X-Cache-Time-Remaining", strconv.Itoa(int(expiresAt.Seconds())))
	}

	return ctx.JSON(response)
}

func IconHandler(ctx *fiber.Ctx) error {
	host, port, err := ParseAddress(ctx.Params("address"), 25565)

	if err != nil {
		return ctx.Status(http.StatusBadRequest).SendString("Invalid address value")
	}

	icon, expiresAt, err := status.GetServerIcon(r, host, port, config.Cache.IconCacheDuration)

	if err != nil {
		return err
	}

	if expiresAt != nil {
		ctx.Set("X-Cache-Time-Remaining", strconv.Itoa(int(expiresAt.Seconds())))
	}

	return ctx.Type("png").Send(icon)
}

func DefaultIconHandler(ctx *fiber.Ctx) error {
	return ctx.Type("png").Send(util.DefaultIconBytes)
}

func NotFoundHandler(ctx *fiber.Ctx) error {
	return ctx.SendStatus(http.StatusNotFound)
}
