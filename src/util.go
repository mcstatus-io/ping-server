package main

import (
	"crypto/sha1"
	_ "embed"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var (
	//go:embed icon.png
	defaultIconBytes []byte
	ipAddressRegExp  = regexp.MustCompile(`^\d{1,3}(\.\d{1,3}){3}$`)
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

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	blockedServers = &MutexArray[string]{
		List:  strings.Split(string(body), "\n"),
		Mutex: &sync.Mutex{},
	}

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

		if blockedServers.Has(newAddressHash) {
			return true
		}
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

func GetInstanceID() (uint16, error) {
	if instanceID := os.Getenv("INSTANCE_ID"); len(instanceID) > 0 {
		value, err := strconv.ParseUint(instanceID, 10, 16)

		if err != nil {
			log.Fatal(err)
		}

		return uint16(value), nil
	}

	return 0, nil
}

type MutexArray[K comparable] struct {
	List  []K
	Mutex *sync.Mutex
}

func (m *MutexArray[K]) Has(value K) bool {
	m.Mutex.Lock()

	defer m.Mutex.Unlock()

	for _, v := range m.List {
		if v == value {
			return true
		}
	}

	return false
}
