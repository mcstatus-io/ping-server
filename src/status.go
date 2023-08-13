package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"main/src/assets"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mcstatus-io/mcutil"
	"github.com/mcstatus-io/mcutil/description"
	"github.com/mcstatus-io/mcutil/options"
	"github.com/mcstatus-io/mcutil/response"
)

// BaseStatus is the base response properties for returning any status response from the API.
type BaseStatus struct {
	Online      bool   `json:"online"`
	Host        string `json:"host"`
	Port        uint16 `json:"port"`
	EULABlocked bool   `json:"eula_blocked"`
	RetrievedAt int64  `json:"retrieved_at"`
	ExpiresAt   int64  `json:"expires_at"`
}

// JavaStatusResponse is the combined response of the root response and the Java Edition status response.
type JavaStatusResponse struct {
	BaseStatus
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
	BaseStatus
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
func GetJavaStatus(host string, port uint16, query bool) (*JavaStatusResponse, time.Duration, error) {
	cacheKey := GetCacheKey(host, port, query)

	// Wait for any other processes to finish fetching the status of this server
	if conf.Cache.EnableLocks {
		mutex := r.NewMutex(fmt.Sprintf("java-lock:%s", cacheKey))
		mutex.Lock()

		defer mutex.Unlock()
	}

	// Fetch the cached status if it exists
	{
		cache, ttl, err := r.Get(fmt.Sprintf("java:%s", cacheKey))

		if err != nil {
			return nil, 0, err
		}

		if cache != nil {
			var response JavaStatusResponse

			err = json.Unmarshal(cache, &response)

			return &response, ttl, err
		}
	}

	// Fetch a fresh status from the server itself
	{
		response := FetchJavaStatus(host, port, query)

		data, err := json.Marshal(response)

		if err != nil {
			return nil, 0, err
		}

		if err := r.Set(fmt.Sprintf("java:%s", cacheKey), data, conf.Cache.JavaStatusDuration); err != nil {
			return nil, 0, err
		}

		return &response, 0, nil
	}
}

// GetBedrockStatus returns the status response of a Bedrock Edition server, either using cache or fetching a fresh status.
func GetBedrockStatus(host string, port uint16) (*BedrockStatusResponse, time.Duration, error) {
	cacheKey := GetCacheKey(host, port, false)

	// Wait for any other processes to finish fetching the status of this server
	if conf.Cache.EnableLocks {
		mutex := r.NewMutex(fmt.Sprintf("bedrock-lock:%s", cacheKey))
		mutex.Lock()

		defer mutex.Unlock()
	}

	// Fetch the cached status if it exists
	{
		cache, ttl, err := r.Get(fmt.Sprintf("bedrock:%s", cacheKey))

		if err != nil {
			return nil, 0, err
		}

		if cache != nil {
			var response BedrockStatusResponse

			err = json.Unmarshal(cache, &response)

			return &response, ttl, err
		}
	}

	var (
		err      error                  = nil
		response *BedrockStatusResponse = nil
		data     []byte                 = nil
	)

	// Fetch a fresh status from the server itself
	{
		response = FetchBedrockStatus(host, port)

		if data, err = json.Marshal(response); err != nil {
			return nil, 0, err
		}
	}

	// Put the status into the cache for future requests
	if err = r.Set(fmt.Sprintf("bedrock:%s", cacheKey), data, conf.Cache.BedrockStatusDuration); err != nil {
		return nil, 0, err
	}

	return response, 0, nil
}

// GetServerIcon returns the icon image of a Java Edition server, either using cache or fetching a fresh image.
func GetServerIcon(host string, port uint16) ([]byte, time.Duration, error) {
	cacheKey := GetCacheKey(host, port, false)

	// Fetch the cached icon if it exists
	{
		cache, ttl, err := r.Get(fmt.Sprintf("icon:%s", cacheKey))

		if err != nil {
			return nil, 0, err
		}

		if cache != nil {
			return cache, ttl, err
		}
	}

	var (
		icon []byte = nil
	)

	// Fetch the icon from the server itself
	{
		status, err := mcutil.Status(host, port)

		if err == nil && status.Favicon != nil && strings.HasPrefix(*status.Favicon, "data:image/png;base64,") {
			data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(*status.Favicon, "data:image/png;base64,"))

			if err != nil {
				return nil, 0, err
			}

			icon = data
		} else {
			icon = assets.DefaultIcon
		}
	}

	// Put the icon into the cache for future requests
	if err := r.Set(fmt.Sprintf("icon:%s", cacheKey), icon, conf.Cache.IconDuration); err != nil {
		return nil, 0, err
	}

	return icon, 0, nil
}

// FetchJavaStatus fetches fresh information about a Java Edition Minecraft server.
func FetchJavaStatus(host string, port uint16, enableQuery bool) JavaStatusResponse {
	var wg sync.WaitGroup

	wg.Add(1)

	if enableQuery {
		wg.Add(1)
	}

	var (
		statusResult interface{}         = nil
		queryResult  *response.FullQuery = nil
	)

	// Status
	{
		go func() {
			if result, _ := mcutil.Status(host, port); result != nil {
				statusResult = result
			} else if result, _ := mcutil.StatusLegacy(host, port); result != nil {
				statusResult = result
			}

			wg.Done()
		}()
	}

	// Query
	if enableQuery {
		go func() {
			queryResult, _ = mcutil.FullQuery(host, port, options.Query{
				Timeout: time.Second,
			})

			wg.Done()
		}()
	}

	wg.Wait()

	return BuildJavaResponse(host, port, statusResult, queryResult)
}

// FetchBedrockStatus fetches a fresh status of a Bedrock Edition server.
func FetchBedrockStatus(host string, port uint16) *BedrockStatusResponse {
	status, err := mcutil.StatusBedrock(host, port)

	if err != nil {
		return &BedrockStatusResponse{
			BaseStatus: BaseStatus{
				Online:      false,
				Host:        host,
				Port:        port,
				EULABlocked: IsBlockedAddress(host),
				RetrievedAt: time.Now().UnixMilli(),
				ExpiresAt:   time.Now().Add(conf.Cache.BedrockStatusDuration).UnixMilli(),
			},
		}
	}

	response := &BedrockStatusResponse{
		BaseStatus: BaseStatus{
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

	return response
}

// BuildJavaResponse builds the response data from the status and query information.
func BuildJavaResponse(host string, port uint16, status interface{}, query *response.FullQuery) (result JavaStatusResponse) {
	result = JavaStatusResponse{
		BaseStatus: BaseStatus{
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

			for _, username := range query.Players {
				parsedName, err := description.ParseFormatting(username)

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

		for _, username := range query.Players {
			if Contains(Map(result.Players.List, func(v Player) string { return v.NameRaw }), username) {
				continue
			}

			parsedName, err := description.ParseFormatting(username)

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

	return
}
