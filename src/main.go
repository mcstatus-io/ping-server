package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
)

var (
	app *fiber.App = fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			var fiberError *fiber.Error

			if errors.As(err, &fiberError) {
				return ctx.SendStatus(fiberError.Code)
			}

			log.Printf("Error: %v - URI: %s\n", err, ctx.Request().URI())

			return ctx.SendStatus(http.StatusInternalServerError)
		},
	})
	r          *Redis  = &Redis{}
	conf       *Config = DefaultConfig
	instanceID uint16  = 0
)

func init() {
	var err error

	if err = conf.ReadFile("config.yml"); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("config.yml does not exist, writing default config\n")

			if err = conf.WriteFile("config.yml"); err != nil {
				log.Fatalf("Failed to write config file: %v", err)
			}
		} else {
			log.Printf("Failed to read config file: %v", err)
		}
	}

	if err = GetBlockedServerList(); err != nil {
		log.Fatalf("Failed to retrieve EULA blocked servers: %v", err)
	}

	log.Println("Successfully retrieved EULA blocked servers")

	if conf.Redis != nil {
		if err = r.Connect(); err != nil {
			log.Fatalf("Failed to connect to Redis: %v", err)
		}

		log.Println("Successfully connected to Redis")
	}

	if instanceID, err = GetInstanceID(); err != nil {
		panic(err)
	}

	app.Hooks().OnListen(func(ld fiber.ListenData) error {
		log.Printf("Listening on %s:%d\n", conf.Host, conf.Port+instanceID)

		return nil
	})
}

func main() {
	defer r.Close()

	if err := app.Listen(fmt.Sprintf("%s:%d", conf.Host, conf.Port+instanceID)); err != nil {
		panic(err)
	}
}
