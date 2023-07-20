package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mcstatus-io/mcutil"
	"github.com/mcstatus-io/mcutil/description"
	"github.com/mcstatus-io/mcutil/options"
	"github.com/mcstatus-io/mcutil/response"
)

// StatusResponse is the root response for returning any status response from the API.
type StatusResponse struct {
	Online      bool   `json:"online"`
	Host        string `json:"host"`
	Port        uint16 `json:"port"`
	EULABlocked bool   `json:"eula_blocked"`
	RetrievedAt int64  `json:"retrieved_at"`
	ExpiresAt   int64  `json:"expires_at"`
}

// JavaStatusResponse is the combined response of the root response and the Java Edition status response.
type JavaStatusResponse struct {
	StatusResponse
	*JavaStatus
}

// JavaStatus is the status response properties for Java Edition.
type JavaStatus struct {
	Version  *JavaVersion `json:"version"`
	Players  JavaPlayers  `json:"players"`
	MOTD     MOTD         `json:"motd"`
	Icon     *string      `json:"icon"`
	Mods     []Mod        `json:"mods"`
	Software *string      `json:"software"`
	Plugins  []Plugin     `json:"plugins"`
}

// BedrockStatusResponse is the combined response of the root response and the Bedrock Edition status response.
type BedrockStatusResponse struct {
	StatusResponse
	*BedrockStatus
}

// BedrockStatus is the status response properties for Bedrock Edition.
type BedrockStatus struct {
	Version  *BedrockVersion `json:"version"`
	Players  *BedrockPlayers `json:"players"`
	MOTD     *MOTD           `json:"motd"`
	Gamemode *string         `json:"gamemode"`
	ServerID *string         `json:"server_id"`
	Edition  *string         `json:"edition"`
}

// JavaVersion holds the properties for the version of Java Edition responses.
type JavaVersion struct {
	NameRaw   string `json:"name_raw"`
	NameClean string `json:"name_clean"`
	NameHTML  string `json:"name_html"`
	Protocol  int64  `json:"protocol"`
}

// BedrockVersion holds the properties for the version of Bedrock Edition responses.
type BedrockVersion struct {
	Name     *string `json:"name"`
	Protocol *int64  `json:"protocol"`
}

// JavaPlayers holds the properties for the players of Java Edition responses.
type JavaPlayers struct {
	Online *int64   `json:"online"`
	Max    *int64   `json:"max"`
	List   []Player `json:"list"`
}

// BedrockPlayers holds the properties for the players of Bedrock Edition responses.
type BedrockPlayers struct {
	Online *int64 `json:"online"`
	Max    *int64 `json:"max"`
}

// Player is a single sample player used in Java Edition status responses.
type Player struct {
	UUID      string `json:"uuid"`
	NameRaw   string `json:"name_raw"`
	NameClean string `json:"name_clean"`
	NameHTML  string `json:"name_html"`
}

// MOTD is a group of formatted and unformatted properties for status responses.
type MOTD struct {
	Raw   string `json:"raw"`
	Clean string `json:"clean"`
	HTML  string `json:"html"`
}

// Mod is a single Forge mod installed on any Java Edition status response.
type Mod struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Plugin is a plugin that is enabled on a Java Edition server.
type Plugin struct {
	Name    string  `json:"name"`
	Version *string `json:"version"`
}

// GetJavaStatus returns the status response of a Java Edition server, either using cache or fetching a fresh status.
func GetJavaStatus(host string, port uint16, enableQuery bool) (*JavaStatusResponse, time.Duration, error) {
	cacheKey := fmt.Sprintf("java:%v-%s-%d", enableQuery, host, port)

	mutex := r.NewMutex(fmt.Sprintf("java-lock:%v-%s-%d", enableQuery, host, port))
	mutex.Lock()

	defer mutex.Unlock()

	cache, ttl, err := r.Get(cacheKey)

	if err != nil {
		return nil, 0, err
	}

	if cache != nil {
		var response JavaStatusResponse

		err = json.Unmarshal(cache, &response)

		return &response, ttl, err
	}

	response := FetchJavaStatus(host, port, enableQuery)

	data, err := json.Marshal(response)

	if err != nil {
		return nil, 0, err
	}

	if err := r.Set(cacheKey, data, conf.Cache.JavaStatusDuration); err != nil {
		return nil, 0, err
	}

	return &response, 0, nil
}

// GetBedrockStatus returns the status response of a Bedrock Edition server, either using cache or fetching a fresh status.
func GetBedrockStatus(host string, port uint16) (*BedrockStatusResponse, time.Duration, error) {
	cacheKey := fmt.Sprintf("bedrock:%s-%d", host, port)

	mutex := r.NewMutex(fmt.Sprintf("bedrock-lock:%s-%d", host, port))
	mutex.Lock()

	defer mutex.Unlock()

	cache, ttl, err := r.Get(cacheKey)

	if err != nil {
		return nil, 0, err
	}

	if cache != nil {
		var response BedrockStatusResponse

		err = json.Unmarshal(cache, &response)

		return &response, ttl, err
	}

	response, err := FetchBedrockStatus(host, port)

	if err != nil {
		return nil, 0, err
	}

	data, err := json.Marshal(response)

	if err != nil {
		return nil, 0, err
	}

	if err := r.Set(cacheKey, data, conf.Cache.BedrockStatusDuration); err != nil {
		return nil, 0, err
	}

	return response, 0, nil
}

// GetServerIcon returns the icon image of a Java Edition server, either using cache or fetching a fresh image.
func GetServerIcon(host string, port uint16) ([]byte, time.Duration, error) {
	cacheKey := fmt.Sprintf("icon:%s-%d", host, port)

	cache, ttl, err := r.Get(cacheKey)

	if err != nil {
		return nil, 0, err
	}

	if cache != nil {
		return cache, ttl, err
	}

	icon := defaultIconBytes

	status, err := mcutil.Status(host, port)

	if err == nil && status.Favicon != nil && strings.HasPrefix(*status.Favicon, "data:image/png;base64,") {
		data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(*status.Favicon, "data:image/png;base64,"))

		if err != nil {
			return nil, 0, err
		}

		icon = data
	}

	if err := r.Set(cacheKey, icon, conf.Cache.IconDuration); err != nil {
		return nil, 0, err
	}

	return icon, 0, nil
}

// FetchJavaStatus fetches fresh information about a Java Edition Minecraft server.
func FetchJavaStatus(host string, port uint16, enableQuery bool) JavaStatusResponse {
	var wg sync.WaitGroup

	var status interface{} = nil
	var query *response.FullQuery = nil

	go func() {
		defer wg.Done()

		if result, _ := mcutil.Status(host, port); result != nil {
			status = result
		} else if result, _ := mcutil.StatusLegacy(host, port); result != nil {
			status = result
		}
	}()

	wg.Add(1)

	if enableQuery {
		go func() {
			defer wg.Done()

			query, _ = mcutil.FullQuery(host, port, options.Query{
				Timeout: time.Second,
			})
		}()

		wg.Add(1)
	}

	wg.Wait()

	return BuildJavaResponse(host, port, status, query)
}

// FetchBedrockStatus fetches a fresh status of a Bedrock Edition server.
func FetchBedrockStatus(host string, port uint16) (*BedrockStatusResponse, error) {
	status, err := mcutil.StatusBedrock(host, port)

	if err != nil {
		return &BedrockStatusResponse{
			StatusResponse: StatusResponse{
				Online:      false,
				Host:        host,
				Port:        port,
				EULABlocked: IsBlockedAddress(host),
				RetrievedAt: time.Now().UnixMilli(),
				ExpiresAt:   time.Now().Add(conf.Cache.BedrockStatusDuration).UnixMilli(),
			},
		}, nil
	}

	response := &BedrockStatusResponse{
		StatusResponse: StatusResponse{
			Online:      true,
			Host:        host,
			Port:        port,
			EULABlocked: IsBlockedAddress(host),
			RetrievedAt: time.Now().UnixMilli(),
			ExpiresAt:   time.Now().Add(conf.Cache.BedrockStatusDuration).UnixMilli(),
		},
		BedrockStatus: &BedrockStatus{
			Version:  nil,
			Players:  nil,
			MOTD:     nil,
			Gamemode: status.Gamemode,
			ServerID: status.ServerID,
			Edition:  status.Edition,
		},
	}

	if status.Version != nil {
		if response.Version == nil {
			response.Version = &BedrockVersion{
				Name:     nil,
				Protocol: nil,
			}
		}

		response.Version.Name = status.Version
	}

	if status.ProtocolVersion != nil {
		if response.Version == nil {
			response.Version = &BedrockVersion{
				Name:     nil,
				Protocol: nil,
			}
		}

		response.Version.Protocol = status.ProtocolVersion
	}

	if status.OnlinePlayers != nil {
		if response.Players == nil {
			response.Players = &BedrockPlayers{
				Online: nil,
				Max:    nil,
			}
		}

		response.Players.Online = status.OnlinePlayers
	}

	if status.MaxPlayers != nil {
		if response.Players == nil {
			response.Players = &BedrockPlayers{
				Online: nil,
				Max:    nil,
			}
		}

		response.Players.Max = status.MaxPlayers
	}

	if status.MOTD != nil {
		response.MOTD = &MOTD{
			Raw:   status.MOTD.Raw,
			Clean: status.MOTD.Clean,
			HTML:  status.MOTD.HTML,
		}
	}

	return response, nil
}

// BuildJavaResponse builds the response data from the status and query information.
func BuildJavaResponse(host string, port uint16, status interface{}, query *response.FullQuery) (result JavaStatusResponse) {
	result = JavaStatusResponse{
		StatusResponse: StatusResponse{
			Online:      status != nil || query != nil,
			Host:        host,
			Port:        port,
			EULABlocked: IsBlockedAddress(host),
			RetrievedAt: time.Now().UnixMilli(),
			ExpiresAt:   time.Now().Add(conf.Cache.JavaStatusDuration).UnixMilli(),
		},
		JavaStatus: nil,
	}

	if status == nil && query == nil {
		return
	}

	result.JavaStatus = &JavaStatus{
		Players: JavaPlayers{
			List: make([]Player, 0),
		},
		Mods:    make([]Mod, 0),
		Plugins: make([]Plugin, 0),
	}

	if status != nil {
		switch s := status.(type) {
		case *response.JavaStatus:
			{
				result.Version = &JavaVersion{
					NameRaw:   s.Version.NameRaw,
					NameClean: s.Version.NameClean,
					NameHTML:  s.Version.NameHTML,
					Protocol:  s.Version.Protocol,
				}

				result.Players = JavaPlayers{
					Online: s.Players.Online,
					Max:    s.Players.Max,
					List:   make([]Player, 0),
				}

				result.MOTD = MOTD{
					Raw:   s.MOTD.Raw,
					Clean: s.MOTD.Clean,
					HTML:  s.MOTD.HTML,
				}

				result.Icon = s.Favicon

				if s.Players.Sample != nil {
					for _, player := range s.Players.Sample {
						result.Players.List = append(result.Players.List, Player{
							UUID:      player.ID,
							NameRaw:   player.NameRaw,
							NameClean: player.NameClean,
							NameHTML:  player.NameHTML,
						})
					}
				}

				if s.ModInfo != nil {
					for _, mod := range s.ModInfo.Mods {
						result.Mods = append(result.Mods, Mod{
							Name:    mod.ID,
							Version: mod.Version,
						})
					}
				}

				break
			}
		case *response.JavaStatusLegacy:
			{
				if s.Version != nil {
					result.Version = &JavaVersion{
						NameRaw:   s.Version.NameRaw,
						NameClean: s.Version.NameClean,
						NameHTML:  s.Version.NameHTML,
						Protocol:  s.Version.Protocol,
					}
				}

				result.Players = JavaPlayers{
					Online: &s.Players.Online,
					Max:    &s.Players.Max,
					List:   make([]Player, 0),
				}

				result.MOTD = MOTD{
					Raw:   s.MOTD.Raw,
					Clean: s.MOTD.Clean,
					HTML:  s.MOTD.HTML,
				}

				break
			}
		default:
			panic(fmt.Errorf("unknown status type: %T", status))
		}
	}

	if query != nil {
		if status == nil {
			if motd, ok := query.Data["hostname"]; ok {
				if parsedMOTD, err := description.ParseFormatting(motd); err == nil {
					result.MOTD = MOTD{
						Raw:   parsedMOTD.Raw,
						Clean: parsedMOTD.Clean,
						HTML:  parsedMOTD.HTML,
					}
				}
			}

			if onlinePlayers, ok := query.Data["numplayers"]; ok {
				value, err := strconv.ParseInt(onlinePlayers, 10, 64)

				if err == nil {
					result.Players.Online = &value
				}
			}

			if maxPlayers, ok := query.Data["maxplayers"]; ok {
				value, err := strconv.ParseInt(maxPlayers, 10, 64)

				if err == nil {
					result.Players.Max = &value
				}
			}

			for _, playerName := range query.Players {
				parsedName, err := description.ParseFormatting(playerName)

				if err == nil {
					result.Players.List = append(result.Players.List, Player{
						UUID:      "",
						NameRaw:   parsedName.Raw,
						NameClean: parsedName.Clean,
						NameHTML:  parsedName.HTML,
					})
				}
			}
		}

		if plugins, ok := query.Data["plugins"]; ok {
			if softwareSplit := strings.Split(strings.Trim(plugins, " "), ":"); len(softwareSplit) > 1 {
				result.Software = PointerOf(strings.Trim(softwareSplit[0], " "))

				for _, plugin := range strings.Split(softwareSplit[1], ";") {
					pluginSplit := strings.Split(strings.Trim(plugin, " "), " ")

					if len(pluginSplit) > 1 {
						result.Plugins = append(result.Plugins, Plugin{
							Name:    pluginSplit[0],
							Version: PointerOf(pluginSplit[1]),
						})
					} else {
						result.Plugins = append(result.Plugins, Plugin{
							Name:    pluginSplit[0],
							Version: nil,
						})
					}
				}
			}
		}
	}

	return
}
