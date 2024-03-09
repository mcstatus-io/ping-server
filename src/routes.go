package main

import (
	"context"
	"fmt"
	"main/src/assets"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/mcstatus-io/mcutil/v3"
	"github.com/mcstatus-io/mcutil/v3/options"
)

func init() {
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	app.Use(favicon.New(favicon.Config{
		Data: assets.Favicon,
	}))

	if config.AccessControl.Enable {
		app.Use(cors.New(cors.Config{
			AllowOrigins:  strings.Join(config.AccessControl.AllowedOrigins, ","),
			AllowMethods:  "HEAD,OPTIONS,GET,POST",
			ExposeHeaders: "X-Cache-Hit,X-Cache-Time-Remaining",
		}))
	}

	if config.Environment == "development" {
		app.Use(logger.New(logger.Config{
			Format:     "${time} ${ip}:${port} -> ${status}: ${method} ${path} (${latency})\n",
			TimeFormat: "2006/01/02 15:04:05",
		}))
	}

	app.Get("/ping", PingHandler)
	app.Get("/status/java/:address", JavaStatusHandler)
	app.Get("/status/bedrock/:address", BedrockStatusHandler)
	app.Get("/icon", DefaultIconHandler)
	app.Get("/icon/:address", IconHandler)
	app.Post("/vote", SendVoteHandler)
}

// PingHandler responds with a 200 OK status for simple health checks.
func PingHandler(ctx *fiber.Ctx) error {
	return ctx.SendStatus(http.StatusOK)
}

// JavaStatusHandler returns the status of the Java edition Minecraft server specified in the address parameter.
func JavaStatusHandler(ctx *fiber.Ctx) error {
	opts, err := GetStatusOptions(ctx)

	if err != nil {
		return err
	}

	host, port, err := ParseAddress(ctx.Params("address"), 25565)

	if err != nil {
		return ctx.Status(http.StatusBadRequest).SendString("Invalid address value")
	}

	if err = r.Increment(fmt.Sprintf("java-hits:%s-%d", host, port)); err != nil {
		return err
	}

	response, expiresAt, err := GetJavaStatus(host, port, opts)

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
	opts, err := GetStatusOptions(ctx)

	if err != nil {
		return err
	}

	host, port, err := ParseAddress(ctx.Params("address"), 19132)

	if err != nil {
		return ctx.Status(http.StatusBadRequest).SendString("Invalid address value")
	}

	if err = r.Increment(fmt.Sprintf("bedrock-hits:%s-%d", host, port)); err != nil {
		return err
	}

	response, expiresAt, err := GetBedrockStatus(host, port, opts)

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
	opts, err := GetStatusOptions(ctx)

	if err != nil {
		return err
	}

	host, port, err := ParseAddress(ctx.Params("address"), 25565)

	if err != nil {
		return ctx.Status(http.StatusBadRequest).SendString("Invalid address value")
	}

	icon, expiresAt, err := GetServerIcon(host, port, opts)

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

// SendVoteHandler allows sending of Votifier votes to the specified server.
func SendVoteHandler(ctx *fiber.Ctx) error {
	opts, err := GetVoteOptions(ctx)

	if err != nil {
		return ctx.Status(http.StatusBadRequest).SendString(err.Error())
	}

	switch opts.Version {
	case 1:
		{
			c, cancel := context.WithTimeout(context.Background(), opts.Timeout)

			defer cancel()

			if err = mcutil.SendVote(c, opts.Host, opts.Port, options.Vote{
				PublicKey:   opts.PublicKey,
				ServiceName: opts.ServiceName,
				Username:    opts.Username,
				IPAddress:   opts.IPAddress,
				Timestamp:   opts.Timestamp,
				Timeout:     opts.Timeout,
			}); err != nil {
				return ctx.Status(http.StatusBadRequest).SendString(err.Error())
			}
		}
	case 2:
		{
			c, cancel := context.WithTimeout(context.Background(), opts.Timeout)

			defer cancel()

			if err = mcutil.SendVote(c, opts.Host, opts.Port, options.Vote{
				ServiceName: opts.ServiceName,
				Username:    opts.Username,
				Token:       opts.Token,
				UUID:        opts.UUID,
				Timestamp:   opts.Timestamp,
				Timeout:     opts.Timeout,
			}); err != nil {
				return ctx.Status(http.StatusBadRequest).SendString(err.Error())
			}
		}
	}

	return ctx.Status(http.StatusOK).SendString("The vote was successfully sent to the server")
}
