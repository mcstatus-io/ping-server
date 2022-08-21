package main

import (
	"crypto/sha1"
	_ "embed"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var (
	//go:embed icon.png
	defaultIconBytes []byte
	ipAddressRegExp  = regexp.MustCompile(`^\d{1,3}(\.\d{1,3}){3}$`)
)

func Contains[T comparable](arr []T, v T) bool {
	for _, value := range arr {
		if v == value {
			return true
		}
	}

	return false
}

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
		return "", 0, fmt.Errorf("'%s' does not match any known address", address)
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
