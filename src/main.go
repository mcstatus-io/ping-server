package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

var (
	app *fiber.App = fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			log.Println(err)

			return ctx.SendStatus(http.StatusInternalServerError)
		},
	})
	r              *Redis              = &Redis{}
	config         *Config             = &Config{}
	blockedServers *MutexArray[string] = nil
)

func init() {
	if err := config.ReadFile("config.yml"); err != nil {
		log.Fatal(err)
	}

	if err := r.Connect(config.Redis); err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully connected to Redis")

	if err := GetBlockedServerList(); err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully retrieved EULA blocked servers")

	app.Config()
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:  "*",
		AllowMethods:  "HEAD,OPTIONS,GET",
		ExposeHeaders: "Content-Type,X-Cache-Time-Remaining",
	}))
}

func main() {
	instanceID, err := GetInstanceID()

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening on %s:%d\n", config.Host, config.Port+instanceID)
	log.Fatal(app.Listen(fmt.Sprintf("%s:%d", config.Host, config.Port+instanceID)))
}
