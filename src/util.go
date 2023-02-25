package main

import (
	"crypto/sha1"
	_ "embed"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	//go:embed icon.png
	defaultIcon     []byte
	ipAddressRegExp *regexp.Regexp = regexp.MustCompile(`^\d{1,3}(\.\d{1,3}){3}$`)
)

func IsBlockedAddress(address string) (bool, error) {
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

		exists, err := r.Exists(fmt.Sprintf("blocked:%s", hex.EncodeToString(newAddressBytes[:])))

		if err != nil {
			return false, err
		}

		if exists {
			return true, nil
		}
	}

	return false, nil
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
