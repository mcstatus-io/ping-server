package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
)

func PingHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetBodyString(http.StatusText(http.StatusOK))
}

func JavaStatusHandler(ctx *fasthttp.RequestCtx) {
	host, port, err := ParseAddress(ctx.UserValue("address").(string), 25565)

	if err != nil {
		WriteError(ctx, nil, http.StatusBadRequest, "Invalid/malformed address")

		return
	}

	cacheEnabled, cacheKey, err := IsCacheEnabled(ctx, "java", host, port)

	if err != nil {
		WriteError(ctx, err, http.StatusInternalServerError)

		return
	}

	if cacheEnabled {
		exists, cache, ttl, err := r.GetValueAndTTL(cacheKey)

		if err != nil {
			WriteError(ctx, err, http.StatusInternalServerError)

			return
		}

		if exists {
			ctx.Response.Header.Set("X-Cache-Hit", "TRUE")
			ctx.Response.Header.Set("X-Cache-Time-Remaining", strconv.Itoa(int(ttl.Seconds())))
			ctx.SetContentType("application/json")
			ctx.SetBodyString(cache)

			return
		}
	}

	ctx.Response.Header.Set("X-Cache-Hit", "FALSE")

	status := GetJavaStatus(host, port)

	if status.Online && status.Response.Favicon != nil {
		data, err := base64.StdEncoding.DecodeString(strings.Replace(*status.Response.Favicon, "data:image/png;base64,", "", 1))

		if err != nil {
			WriteError(ctx, err, http.StatusInternalServerError)

			return
		}

		if err := r.Set(fmt.Sprintf("favicon:%s-%d", host, port), data, config.FaviconCacheTTL); err != nil {
			WriteError(ctx, err, http.StatusInternalServerError)

			return
		}
	}

	data, err := json.Marshal(status)

	if err != nil {
		WriteError(ctx, err, http.StatusInternalServerError)

		return
	}

	if err = r.Set(cacheKey, data, config.StatusCacheTTL); err != nil {
		WriteError(ctx, err, http.StatusInternalServerError)

		return
	}

	ctx.SetContentType("application/json")
	ctx.SetBody(data)
}

func BedrockStatusHandler(ctx *fasthttp.RequestCtx) {
	host, port, err := ParseAddress(ctx.UserValue("address").(string), 19132)

	if err != nil {
		WriteError(ctx, nil, http.StatusBadRequest, "Invalid/malformed address")

		return
	}

	cacheEnabled, cacheKey, err := IsCacheEnabled(ctx, "bedrock", host, port)

	if err != nil {
		WriteError(ctx, err, http.StatusInternalServerError)

		return
	}

	if cacheEnabled {
		exists, cache, ttl, err := r.GetValueAndTTL(cacheKey)

		if err != nil {
			WriteError(ctx, err, http.StatusInternalServerError)

			return
		}

		if exists {
			ctx.Response.Header.Set("X-Cache-Hit", "TRUE")
			ctx.Response.Header.Set("X-Cache-Time-Remaining", strconv.Itoa(int(ttl.Seconds())))
			ctx.SetContentType("application/json")
			ctx.SetBodyString(cache)

			return
		}
	}

	ctx.Response.Header.Set("X-Cache-Hit", "FALSE")

	data, err := json.Marshal(GetBedrockStatus(host, port))

	if err != nil {
		WriteError(ctx, err, http.StatusInternalServerError)

		return
	}

	if err = r.Set(cacheKey, data, config.StatusCacheTTL); err != nil {
		WriteError(ctx, err, http.StatusInternalServerError)

		return
	}

	ctx.SetContentType("application/json")
	ctx.SetBody(data)
}

func FaviconNoExtensionHandler(ctx *fasthttp.RequestCtx) {
	host, port, err := ParseAddress(ctx.UserValue("address").(string), 25565)

	if err != nil {
		WriteError(ctx, nil, http.StatusBadRequest, "Invalid/malformed address")

		return
	}

	cacheKey := fmt.Sprintf("favicon:%s-%d", host, port)

	exists, err := r.Exists(cacheKey)

	if err != nil {
		WriteError(ctx, err, http.StatusInternalServerError)

		return
	}

	if exists {
		value, err := r.GetBytes(cacheKey)

		if err != nil {
			WriteError(ctx, err, http.StatusInternalServerError)

			return
		}

		ctx.Response.Header.Set("X-Cache-Hit", "TRUE")
		ctx.Response.Header.Set("X-Server-Status", "online-cache")
		ctx.SetContentType("image/png")
		ctx.SetBody(value)

		return
	}

	ctx.Response.Header.Set("X-Cache-Hit", "FALSE")

	status := GetJavaStatus(host, port)

	if !status.Online {
		ctx.Response.Header.Set("X-Server-Status", "offline")
		ctx.SetContentType("image/png")
		ctx.SetBody(defaultIconBytes)

		return
	}

	if status.Response.Favicon == nil {
		ctx.Response.Header.Set("X-Server-Status", "online-no-icon")
		ctx.SetContentType("image/png")
		ctx.SetBody(defaultIconBytes)

		return
	}

	data, err := base64.StdEncoding.DecodeString(strings.Replace(*status.Response.Favicon, "data:image/png;base64,", "", 1))

	if err != nil {
		WriteError(ctx, err, http.StatusInternalServerError)

		return
	}

	if err := r.Set(cacheKey, data, config.FaviconCacheTTL); err != nil {
		WriteError(ctx, err, http.StatusInternalServerError)

		return
	}

	ctx.Response.Header.Set("X-Server-Status", "online")
	ctx.SetContentType("image/png")
	ctx.SetBody(data)
}

func FaviconHandler(ctx *fasthttp.RequestCtx) {
	filename := ctx.UserValue("filename").(string)

	if !strings.HasSuffix(filename, ".png") {
		WriteError(ctx, nil, http.StatusBadRequest, "Filename must end with .png")

		return
	}

	host, port, err := ParseAddress(ctx.UserValue("address").(string), 25565)

	if err != nil {
		WriteError(ctx, nil, http.StatusBadRequest, "Invalid/malformed address")

		return
	}

	cacheKey := fmt.Sprintf("favicon:%s-%d", host, port)

	exists, err := r.Exists(cacheKey)

	if err != nil {
		WriteError(ctx, err, http.StatusInternalServerError)

		return
	}

	if exists {
		value, err := r.GetBytes(cacheKey)

		if err != nil {
			WriteError(ctx, err, http.StatusInternalServerError)

			return
		}

		ctx.Response.Header.Set("X-Cache-Hit", "TRUE")
		ctx.Response.Header.Set("X-Server-Status", "online-cache")
		ctx.Response.Header.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))
		ctx.SetContentType("image/png")
		ctx.SetBody(value)

		return
	}

	ctx.Response.Header.Set("X-Cache-Hit", "FALSE")

	status := GetJavaStatus(host, port)

	if !status.Online {
		ctx.Response.Header.Set("X-Server-Status", "offline")
		ctx.Response.Header.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))
		ctx.SetContentType("image/png")
		ctx.SetBody(defaultIconBytes)

		return
	}

	if status.Response.Favicon == nil {
		ctx.Response.Header.Set("X-Server-Status", "online-no-icon")
		ctx.Response.Header.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))
		ctx.SetContentType("image/png")
		ctx.SetBody(defaultIconBytes)

		return
	}

	data, err := base64.StdEncoding.DecodeString(strings.Replace(*status.Response.Favicon, "data:image/png;base64,", "", 1))

	if err != nil {
		WriteError(ctx, err, http.StatusInternalServerError)

		return
	}

	if err := r.Set(cacheKey, data, config.FaviconCacheTTL); err != nil {
		WriteError(ctx, err, http.StatusInternalServerError)

		return
	}

	ctx.Response.Header.Set("X-Server-Status", "online")
	ctx.Response.Header.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))
	ctx.SetContentType("image/png")
	ctx.SetBody(data)
}
