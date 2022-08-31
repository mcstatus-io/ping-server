package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

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
