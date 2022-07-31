package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

var (
	config              *Configuration = &Configuration{}
	r                   *Redis         = &Redis{}
	blockedServers      []string       = nil
	blockedServersMutex *sync.Mutex    = &sync.Mutex{}
)

func init() {
	if err := config.ReadFile("config.yml"); err != nil {
		log.Fatal(err)
	}

	if err := r.Connect(config.Redis); err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully connected to Redis")

	if Contains(os.Args, "--flush-cache") {
		keys, err := r.Keys("java:*")

		if err != nil {
			log.Fatal(err)
		}

		if err = r.Delete(keys...); err != nil {
			log.Fatal(err)
		}

		keys, err = r.Keys("bedrock:*")

		if err != nil {
			log.Fatal(err)
		}

		if err = r.Delete(keys...); err != nil {
			log.Fatal(err)
		}

		keys, err = r.Keys("favicon:*")

		if err != nil {
			log.Fatal(err)
		}

		if err = r.Delete(keys...); err != nil {
			log.Fatal(err)
		}

		log.Println("Successfully flushed all cache keys")

		os.Exit(0)
	}

	if instanceID := os.Getenv("INSTANCE_ID"); len(instanceID) > 0 {
		value, err := strconv.ParseUint(instanceID, 10, 16)

		if err != nil {
			log.Fatal(err)
		}

		config.Port += uint16(value)
	}

	if err := GetBlockedServerList(); err != nil {
		log.Fatal(err)
	}
}

func middleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		if config.Environment == "development" {
			log.Printf("GET %s - %s \"%s\"", ctx.Request.URI().Path(), ctx.RemoteAddr(), ctx.UserAgent())
		}

		ctx.Response.Header.Set("Access-Control-Allow-Headers", "*")
		ctx.Response.Header.Set("Access-Control-Allow-Methods", "HEAD,GET,POST,OPTIONS")
		ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Response.Header.Set("Access-Control-Expose-Headers", "X-Cache-Hit,X-Cache-Time-Remaining,X-Server-Status,Content-Disposition")

		next(ctx)
	}
}

func main() {
	defer r.Close()

	router := fasthttprouter.New()
	router.GET("/ping", PingHandler)
	router.GET("/status/java/:address", JavaStatusHandler)
	router.GET("/status/bedrock/:address", BedrockStatusHandler)
	router.GET("/favicon/:address", FaviconNoExtensionHandler)
	router.GET("/favicon/:address/*filename", FaviconHandler)

	router.PanicHandler = func(rc *fasthttp.RequestCtx, i interface{}) {
		log.Println(i)
	}

	router.NotFound = func(ctx *fasthttp.RequestCtx) {
		WriteError(ctx, nil, http.StatusNotFound)
	}

	log.Printf("Listening on %s:%d\n", config.Host, config.Port)
	log.Fatal(fasthttp.ListenAndServe(fmt.Sprintf("%s:%d", config.Host, config.Port), middleware(router.Handler)))

	s := make(chan os.Signal)
	signal.Notify(s, os.Interrupt, syscall.SIGTERM)
	<-s
}
