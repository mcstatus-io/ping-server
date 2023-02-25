package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

func StartBlockedServersGoroutine() {
	for {
		if err := GetBlockedServerList(); err != nil {
			log.Println(err)
		}

		time.Sleep(time.Hour)
	}
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

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	for _, hash := range strings.Split(string(body), "\n") {
		if err = r.Set(fmt.Sprintf("blocked:%s", hash), "true", time.Hour*24); err != nil {
			return err
		}
	}

	return nil
}
