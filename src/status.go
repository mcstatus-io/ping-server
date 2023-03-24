package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mcstatus-io/mcutil"
)

// StatusResponse contains the common information for a server status response.
type StatusResponse struct {
	Online      bool   `json:"online"`
	Host        string `json:"host"`
	Port        uint16 `json:"port"`
	EULABlocked bool   `json:"eula_blocked"`
	RetrievedAt int64  `json:"retrieved_at"`
	ExpiresAt   int64  `json:"expires_at"`
}

// JavaStatusResponse contains the information for a Java Edition server status response.
type JavaStatusResponse struct {
	StatusResponse
	*JavaStatus
}

// JavaStatus contains the Java Edition specific server status information.
type JavaStatus struct {
	Version *JavaVersion `json:"version"`
	Players JavaPlayers  `json:"players"`
	MOTD    MOTD         `json:"motd"`
	Icon    *string      `json:"icon"`
	Mods    []Mod        `json:"mods"`
}

// BedrockStatusResponse contains the information for a Bedrock Edition server status response.
type BedrockStatusResponse struct {
	StatusResponse
	*BedrockStatus
}

// BedrockStatus contains the Bedrock Edition specific server status information.
type BedrockStatus struct {
	Version  *BedrockVersion `json:"version"`
	Players  *BedrockPlayers `json:"players"`
	MOTD     *MOTD           `json:"motd"`
	Gamemode *string         `json:"gamemode"`
	ServerID *string         `json:"server_id"`
	Edition  *string         `json:"edition"`
}

// JavaVersion contains information about the Java Edition server version.
type JavaVersion struct {
	NameRaw   string `json:"name_raw"`
	NameClean string `json:"name_clean"`
	NameHTML  string `json:"name_html"`
	Protocol  int    `json:"protocol"`
}

// BedrockVersion contains information about the Bedrock Edition server version.
type BedrockVersion struct {
	Name     *string `json:"name"`
	Protocol *int64  `json:"protocol"`
}

// JavaPlayers contains information about the Java Edition server players.
type JavaPlayers struct {
	Online int      `json:"online"`
	Max    int      `json:"max"`
	List   []Player `json:"list"`
}

// BedrockPlayers contains information about the Bedrock Edition server players.
type BedrockPlayers struct {
	Online *int64 `json:"online"`
	Max    *int64 `json:"max"`
}

// Player contains information about a specific player on a Java Edition server.
type Player struct {
	UUID      string `json:"uuid"`
	NameRaw   string `json:"name_raw"`
	NameClean string `json:"name_clean"`
	NameHTML  string `json:"name_html"`
}

// MOTD contains the server's message of the day.
type MOTD struct {
	Raw   string `json:"raw"`
	Clean string `json:"clean"`
	HTML  string `json:"html"`
}

// Mod contains information about a specific mod on a Java Edition server.
type Mod struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// GetJavaStatus retrieves the status of a Java Edition server.
func GetJavaStatus(host string, port uint16) (*JavaStatusResponse, time.Duration, error) {
	cacheKey := fmt.Sprintf("java:%s-%d", host, port)

	cache, ttl, err := r.Get(cacheKey)

	if err != nil {
		return nil, 0, err
	}

	if cache != nil {
		var response JavaStatusResponse
		err = json.Unmarshal(cache, &response)
		return &response, ttl, err
	}

	response, err := fetchJavaStatus(host, port)

	if err != nil {
		return nil, 0, err
	}

	data, err := json.Marshal(response)

	if err != nil {
		return nil, 0, err
	}

	if err := r.Set(cacheKey, data, config.Cache.JavaStatusDuration); err != nil {
		return nil, 0, err
	}

	return response, 0, nil
}

// GetBedrockStatus retrieves the status of a Bedrock Edition server.
func GetBedrockStatus(host string, port uint16) (*BedrockStatusResponse, time.Duration, error) {
	cacheKey := fmt.Sprintf("bedrock:%s-%d", host, port)

	cache, ttl, err := r.Get(cacheKey)

	if err != nil {
		return nil, 0, err
	}

	if cache != nil {
		var response BedrockStatusResponse
		err = json.Unmarshal(cache, &response)
		return &response, ttl, err
	}

	response, err := fetchBedrockStatus(host, port)

	if err != nil {
		return nil, 0, err
	}

	data, err := json.Marshal(response)

	if err != nil {
		return nil, 0, err
	}

	if err := r.Set(cacheKey, data, config.Cache.BedrockStatusDuration); err != nil {
		return nil, 0, err
	}

	return response, 0, nil
}

// GetServerIcon retrieves the server icon for a Java Edition server.
func GetServerIcon(host string, port uint16) ([]byte, time.Duration, error) {
	cacheKey := fmt.Sprintf("icon:%s-%d", host, port)

	cache, ttl, err := r.Get(cacheKey)

	if err != nil {
		return nil, 0, err
	}

	if cache != nil {
		return cache, ttl, err
	}

	icon := defaultIcon

	status, err := mcutil.Status(host, port)

	if err == nil && status.Favicon != nil && strings.HasPrefix(*status.Favicon, "data:image/png;base64,") {
		data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(*status.Favicon, "data:image/png;base64,"))

		if err != nil {
			return nil, 0, err
		}

		icon = data
	}

	if err := r.Set(cacheKey, icon, config.Cache.IconDuration); err != nil {
		return nil, 0, err
	}

	return icon, 0, nil
}

// fetchJavaStatus fetches the Java Edition server status without using the cache.
func fetchJavaStatus(host string, port uint16) (*JavaStatusResponse, error) {
	status, err := mcutil.Status(host, port)
	if err != nil {
		status, err = mcutil.StatusLegacy(host, port)
		if err != nil {
			return nil, err
		}
	}

	response := &JavaStatusResponse{
		StatusResponse: StatusResponse{
			Online: true,
			Host:   host,
			Port:   port,
		},
		JavaStatus: &JavaStatus{
			MOTD: MOTD{
				Raw:   status.Description.Text,
				Clean: status.Description.Text, // Adjust if you want to remove color codes
				HTML:  "",                       // Adjust if you want to generate HTML from the description
			},
			Players: JavaPlayers{
				Online: status.Players.Online,
				Max:    status.Players.Max,
			},
			Version: &JavaVersion{
				NameRaw:   status.Version.Name,
				NameClean: status.Version.Name, // Adjust if you want to remove color codes
				NameHTML:  "",                  // Adjust if you want to generate HTML from the version name
				Protocol:  status.Version.Protocol,
			},
		},
	}

	if status.Favicon != nil {
		response.JavaStatus.Icon = status.Favicon
	}

	return response, nil
}

// fetchBedrockStatus fetches the Bedrock Edition server status without using the cache.
// fetchBedrockStatus fetches the Bedrock Edition server status without using the cache.
func fetchBedrockStatus(host string, port uint16) (*BedrockStatusResponse, error) {
	status, err := mcutil.StatusBedrock(host, port)
	if err != nil {
		return nil, err
	}

	response := &BedrockStatusResponse{
		StatusResponse: StatusResponse{
			Online: true,
			Host:   host,
			Port:   port,
		},
		BedrockStatus: &BedrockStatus{
			Version: &BedrockVersion{
				Name:     status.Version,
				Protocol: status.ProtocolVersion,
			},
			Players: &BedrockPlayers{
				Online: status.Players.Online,
				Max:    status.Players.Max,
			},
			MOTD: &MOTD{
				Raw:   status.Motd,
				Clean: status.Motd, // Adjust if you want to remove color codes
				HTML:  "",          // Adjust if you want to generate HTML from the motd
			},
		},
	}

	return response, nil
}
