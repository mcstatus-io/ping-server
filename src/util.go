package main

import (
	"crypto/sha1"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	//go:embed icon.png
	defaultIconBytes []byte
	ipAddressRegExp  = regexp.MustCompile(`^\d{1,3}(\.\d{1,3}){3}$`)
)

type BlockedServersFile struct {
	LastUpdated time.Time `json:"last_updated"`
	Servers     []string  `json:"servers"`
}

func Contains[T comparable](arr []T, v T) bool {
	for _, value := range arr {
		if v == value {
			return true
		}
	}

	return false
}

func GetBlockedServerList() error {
	f, err := os.OpenFile("blocked-servers.json", os.O_CREATE|os.O_RDWR, 0777)

	if err != nil {
		return err
	}

	defer f.Close()

	data, err := ioutil.ReadAll(f)

	if err != nil {
		return err
	}

	if len(data) > 0 {
		var blockedServersFile BlockedServersFile

		if err = json.Unmarshal(data, &blockedServersFile); err != nil {
			return err
		}

		if time.Since(blockedServersFile.LastUpdated).Hours() < 24 {
			blockedServersMutex.Lock()
			blockedServers = blockedServersFile.Servers
			blockedServersMutex.Unlock()

			return nil
		}
	}

	resp, err := http.Get("https://sessionserver.mojang.com/blockedservers")

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	log.Println("Successfully retrieved EULA blocked servers")

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	blockedServersMutex.Lock()
	blockedServers = strings.Split(string(body), "\n")

	defer blockedServersMutex.Unlock()

	if data, err = json.Marshal(BlockedServersFile{
		LastUpdated: time.Now(),
		Servers:     blockedServers,
	}); err != nil {
		return err
	}

	if err = f.Truncate(0); err != nil {
		return err
	}

	if _, err = f.Seek(0, 0); err != nil {
		return err
	}

	_, err = f.Write(data)

	return err
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
