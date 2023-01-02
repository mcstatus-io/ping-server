package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

var (
	app *fiber.App = fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			log.Println(ctx.Request().URI(), err)

			return ctx.SendStatus(http.StatusInternalServerError)
		},
	})
	r      *Redis  = &Redis{}
	config *Config = &Config{}
)

func init() {
	if err := config.ReadFile("config.yml"); err != nil {
		log.Fatal(err)
	}

	if config.Cache.Enable {
		if err := r.Connect(config.Redis); err != nil {
			log.Fatal(err)
		}

		log.Println("Successfully connected to Redis")
	}

	if err := GetBlockedServerList(); err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully retrieved EULA blocked servers")

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:  "*",
		AllowMethods:  "HEAD,OPTIONS,GET",
		ExposeHeaders: "Content-Type,X-Cache-Time-Remaining",
	}))

	if config.Environment == "development" {
		app.Use(logger.New(logger.Config{
			Format:     "${time} ${ip}:${port} -> ${method} ${path} -> ${status}\n",
			TimeFormat: "2006/01/02 15:04:05",
		}))
	}
}

func main() {
	defer r.Close()

	log.Printf("Listening on %s:%d\n", config.Host, config.Port)
	log.Fatal(app.Listen(fmt.Sprintf("%s:%d", config.Host, config.Port)))
}
