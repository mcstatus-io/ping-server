package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mcstatus-io/mcutil"
	"github.com/mcstatus-io/mcutil/options"
)

type SendVoteBody struct {
	Host     string `json:"host"`
	Port     uint16 `json:"port"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

type SendVoteResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func init() {
	app.Get("/ping", PingHandler)
	app.Get("/status/java/:address", JavaStatusHandler)
	app.Get("/status/bedrock/:address", BedrockStatusHandler)
	app.Get("/icon/default", DefaultIconHandler)
	app.Get("/icon/:address", IconHandler)
	app.Post("/vote", SendVoteHandler)
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

	if err = r.Increment(fmt.Sprintf("java-hits:%s-%d", host, port)); err != nil {
		return err
	}

	response, expiresAt, err := GetJavaStatus(host, port)

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

	if err = r.Increment(fmt.Sprintf("bedrock-hits:%s-%d", host, port)); err != nil {
		return err
	}

	response, expiresAt, err := GetBedrockStatus(host, port)

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

	icon, expiresAt, err := GetServerIcon(host, port)

	if err != nil {
		return err
	}

	if expiresAt != nil {
		ctx.Set("X-Cache-Time-Remaining", strconv.Itoa(int(expiresAt.Seconds())))
	}

	return ctx.Type("png").Send(icon)
}

func DefaultIconHandler(ctx *fiber.Ctx) error {
	return ctx.Type("png").Send(defaultIconBytes)
}

func SendVoteHandler(ctx *fiber.Ctx) error {
	var body SendVoteBody

	if err := ctx.BodyParser(&body); err != nil {
		return err
	}

	if err := mcutil.SendVote(body.Host, body.Port, options.Vote{
		ServiceName: "mcstatus.io Vote Tester",
		Username:    body.Username,
		Token:       body.Token,
		UUID:        "",
		Timestamp:   time.Now(),
		Timeout:     time.Second * 5,
	}); err != nil {
		return ctx.JSON(SendVoteResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return ctx.JSON(SendVoteResponse{
		Success: true,
	})
}

func NotFoundHandler(ctx *fiber.Ctx) error {
	return ctx.SendStatus(http.StatusNotFound)
}
