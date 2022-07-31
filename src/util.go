package main

import (
	"crypto/sha1"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
)

var (
	//go:embed default-icon.png
	defaultIconBytes []byte
	ipAddressRegExp  = regexp.MustCompile("^\\d{1,3}(\\.\\d{1,3}){3}$")
)

var (
	ErrNoAddressMatch = errors.New("address does not match any known format")
)

func GetBlockedServerList() error {
	resp, err := http.Get("https://sessionserver.mojang.com/blockedservers")

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	blockedServersMutex.Lock()
	blockedServers = strings.Split(string(body), "\n")
	blockedServersMutex.Unlock()

	return nil
}

func IsBlockedAddress(address string) bool {
	split := strings.Split(strings.ToLower(address), ".")

	isIPAddress := ipAddressRegExp.MatchString(address)

	for k := range split {
		newAddress := ""

		switch k {
		case 0:
			{
				newAddress = strings.Join(split, ".")

				break
			}
		default:
			{
				if isIPAddress {
					newAddress = fmt.Sprintf("%s.*", strings.Join(split[0:len(split)-k], "."))
				} else {
					newAddress = fmt.Sprintf("*.%s", strings.Join(split[k:], "."))
				}

				break
			}
		}

		newAddressBytes := sha1.Sum([]byte(newAddress))
		newAddressHash := hex.EncodeToString(newAddressBytes[:])

		blockedServersMutex.Lock()

		if Contains(blockedServers, newAddressHash) {
			blockedServersMutex.Unlock()

			return true
		}

		blockedServersMutex.Unlock()
	}

	return false
}

func ParseAddress(address string, defaultPort uint16) (string, uint16, error) {
	result := strings.SplitN(address, ":", 2)

	if len(result) < 1 {
		return "", 0, ErrNoAddressMatch
	}

	if len(result) < 2 {
		return result[0], defaultPort, nil
	}

	port, err := strconv.ParseUint(result[1], 10, 16)

	if err != nil {
		return "", 0, err
	}

	return result[0], uint16(port), nil
}

func WriteError(ctx *fasthttp.RequestCtx, err error, statusCode int, body ...string) {
	ctx.SetStatusCode(statusCode)

	if len(body) > 0 {
		ctx.SetBodyString(body[0])
	} else {
		ctx.SetBodyString(http.StatusText(statusCode))
	}

	if err != nil {
		log.Println(err)
	}
}

func IsCacheEnabled(ctx *fasthttp.RequestCtx, cacheType, host string, port uint16) (bool, string, error) {
	key := fmt.Sprintf("%s:%s-%d", cacheType, host, port)

	if authKey := ctx.Request.Header.Peek("Authorization"); len(authKey) > 0 {
		exists, err := r.Exists(fmt.Sprintf("auth_key:%s", authKey))

		if err != nil {
			WriteError(ctx, err, http.StatusInternalServerError)

			return config.CacheEnable, key, err
		}

		if exists {
			return false, key, nil
		}
	}

	return config.CacheEnable, key, nil
}

func Contains[T comparable](arr []T, x T) bool {
	for _, v := range arr {
		if v == x {
			return true
		}
	}

	return false
}
