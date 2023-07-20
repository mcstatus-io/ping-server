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
	//go:embed favicon.ico
	favicon        []byte
	blockedServers *MutexArray[string] = nil
	ipAddressRegex *regexp.Regexp      = regexp.MustCompile(`^\d{1,3}(\.\d{1,3}){3}$`)
)

// MutexArray is a thread-safe array for storing and retrieving values.
type MutexArray[T comparable] struct {
	List  []T
	Mutex *sync.Mutex
}

// Has checks if the given value is present in the array.
func (m *MutexArray[T]) Has(value T) bool {
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
		return fmt.Errorf("mojang: unexpected status code: %d", resp.StatusCode)
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

// IsBlockedAddress checks if the given address is in the blocked servers list.
func IsBlockedAddress(address string) bool {
	addressSegments := strings.Split(strings.ToLower(address), ".")
	isIPv4Address := ipAddressRegex.MatchString(address)

	for i := range addressSegments {
		var checkAddress string

		if i == 0 {
			checkAddress = strings.Join(addressSegments, ".")
		} else if isIPv4Address {
			checkAddress = fmt.Sprintf("%s.*", strings.Join(addressSegments[0:len(addressSegments)-i], "."))
		} else {
			checkAddress = fmt.Sprintf("*.%s", strings.Join(addressSegments[i:], "."))
		}

		if blockedServers.Has(SHA256(checkAddress)) {
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

// GetInstanceID returns the INSTANCE_ID environment variable parsed as an unsigned 16-bit integer.
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

// SHA256 returns the result of hashing the input value using SHA256 algorithm.
func SHA256(input string) string {
	result := sha1.Sum([]byte(input))

	return hex.EncodeToString(result[:])
}

// PointerOf returns a pointer of the argument passed.
func PointerOf[T any](v T) *T {
	return &v
}
