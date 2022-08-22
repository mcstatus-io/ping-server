package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

var (
	app                 *fiber.App  = nil
	r                   *Redis      = &Redis{}
	config              *Config     = &Config{}
	blockedServers      []string    = nil
	blockedServersMutex *sync.Mutex = &sync.Mutex{}
)

func init() {
	app = fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			log.Println(err)

			return ctx.SendStatus(http.StatusInternalServerError)
		},
	})

	if err := config.ReadFile("config.yml"); err != nil {
		log.Fatal(err)
	}

	r.SetEnabled(config.Cache.Enable)

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

	if instanceID := os.Getenv("INSTANCE_ID"); len(instanceID) > 0 {
		value, err := strconv.ParseUint(instanceID, 10, 16)

		if err != nil {
			log.Fatal(err)
		}

		config.Port += uint16(value)
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:  "*",
		AllowMethods:  "HEAD,OPTIONS,GET",
		ExposeHeaders: "Content-Type,X-Cache-Time-Remaining",
	}))

	app.Use(recover.New())
}

func main() {
	log.Printf("Listening on %s:%d\n", config.Host, config.Port)

	log.Fatal(app.Listen(fmt.Sprintf("%s:%d", config.Host, config.Port)))
}
