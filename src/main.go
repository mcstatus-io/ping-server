package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

var (
	app                 *fiber.App  = fiber.New(fiber.Config{DisableStartupMessage: true})
	r                   *Redis      = &Redis{}
	config              *Config     = &Config{}
	blockedServers      []string    = nil
	blockedServersMutex *sync.Mutex = &sync.Mutex{}
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
		ExposeHeaders: "Content-Type",
	}))
}

func main() {
	log.Printf("Listening on %s:%d\n", config.Host, config.Port)

	log.Fatal(app.Listen(fmt.Sprintf("%s:%d", config.Host, config.Port)))
}
