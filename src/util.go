package main

import (
	"crypto/sha1"
	_ "embed"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var (
	//go:embed icon.png
	defaultIcon    []byte
	blockedServers *MutexArray = nil
	ipAddressRegex *regexp.Regexp = regexp.MustCompile(`^\d{1,3}(\.\d{1,3}){3}$`)
)

// MutexArray is a thread-safe array for storing and checking values.
type MutexArray struct {
	List  []interface{}
	Mutex *sync.Mutex
}

// Has checks if the given value is present in the array.
func (m *MutexArray) Has(value interface{}) bool {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	for _, v := range m.List {
		if v == value {
			return true
		}
	}

	return false
}



// GetBlockedServerList fetches the list of blocked servers from Mojang's session server.
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

	blockedServers = &MutexArray{
		List:  strings.Split(string(body), "\n"),
		Mutex: &sync.Mutex{},
	}

	return nil
}


// IsBlockedAddress checks if the given address is in the blocked servers list.
func IsBlockedAddress(address string) bool {
	split := strings.Split(strings.ToLower(address), ".")
	isIPAddress := ipAddressRegex.MatchString(address)

	for k := range split {
		var newAddress string

		switch k {
		case 0:
			newAddress = strings.Join(split, ".")
		default:
			if isIPAddress {
				newAddress = fmt.Sprintf("%s.*", strings.Join(split[0:len(split)-k], "."))
			} else {
				newAddress = fmt.Sprintf("*.%s", strings.Join(split[k:], "."))
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

// ParseAddress extracts the hostname and port from the given address string, and returns the default port if none is provided.
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
