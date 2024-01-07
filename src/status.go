package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"main/src/assets"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mcstatus-io/mcutil/v3"
	"github.com/mcstatus-io/mcutil/v3/formatting"
	"github.com/mcstatus-io/mcutil/v3/options"
	"github.com/mcstatus-io/mcutil/v3/response"
)

// BaseStatus is the base response properties for returning any status response from the API.
type BaseStatus struct {
	Online      bool    `json:"online"`
	Host        string  `json:"host"`
	Port        uint16  `json:"port"`
	IPAddress   *string `json:"ip_address"`
	EULABlocked bool    `json:"eula_blocked"`
	RetrievedAt int64   `json:"retrieved_at"`
	ExpiresAt   int64   `json:"expires_at"`
}

// JavaStatusResponse is the combined response of the root response and the Java Edition status response.
type JavaStatusResponse struct {
	BaseStatus
	SRVRecord *SRVRecord `json:"srv_record"`
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

// SRVRecord is the result of the SRV lookup performed during status retrieval
type SRVRecord struct {
	Host string `json:"host"`
	Port uint16 `json:"port"`
}

// GetJavaStatus returns the status response of a Java Edition server, either using cache or fetching a fresh status.
func GetJavaStatus(host string, port uint16, opts *StatusOptions) (*JavaStatusResponse, time.Duration, error) {
	cacheKey := GetCacheKey(host, port, opts)

	// Wait for any other processes to finish fetching the status of this server
	if config.Cache.EnableLocks {
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
		response := FetchJavaStatus(host, port, opts)

		data, err := json.Marshal(response)

		if err != nil {
			return nil, 0, err
		}

		if err := r.Set(fmt.Sprintf("java:%s", cacheKey), data, config.Cache.JavaStatusDuration); err != nil {
			return nil, 0, err
		}

		return &response, 0, nil
	}
}

// GetBedrockStatus returns the status response of a Bedrock Edition server, either using cache or fetching a fresh status.
func GetBedrockStatus(host string, port uint16, opts *StatusOptions) (*BedrockStatusResponse, time.Duration, error) {
	cacheKey := GetCacheKey(host, port, nil)

	// Wait for any other processes to finish fetching the status of this server
	if config.Cache.EnableLocks {
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

	// Fetch a fresh status from the server itself
	{
		response := FetchBedrockStatus(host, port, opts)

		data, err := json.Marshal(response)

		if err != nil {
			return nil, 0, err
		}

		if err = r.Set(fmt.Sprintf("bedrock:%s", cacheKey), data, config.Cache.BedrockStatusDuration); err != nil {
			return nil, 0, err
		}

		return &response, 0, nil
	}
}

// GetServerIcon returns the icon image of a Java Edition server, either using cache or fetching a fresh image.
func GetServerIcon(host string, port uint16, opts *StatusOptions) ([]byte, time.Duration, error) {
	cacheKey := GetCacheKey(host, port, nil)

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
		ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)

		defer cancel()

		status, err := mcutil.Status(ctx, host, port)

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
	if err := r.Set(fmt.Sprintf("icon:%s", cacheKey), icon, config.Cache.IconDuration); err != nil {
		return nil, 0, err
	}

	return icon, 0, nil
}

// FetchJavaStatus fetches fresh information about a Java Edition Minecraft server.
func FetchJavaStatus(host string, port uint16, opts *StatusOptions) JavaStatusResponse {
	var (
		err                error
		srvRecord          *net.SRV
		resolvedHost       string = host
		ipAddress          *string
		statusResult       *response.JavaStatus
		legacyStatusResult *response.JavaStatusLegacy
		queryResult        *response.FullQuery
		wg                 sync.WaitGroup
	)

	// Setup initial wait group deltas
	{
		wg.Add(2)

		if opts.Query {
			wg.Add(1)
		}
	}

	// Lookup the SRV record
	{
		srvRecord, err = mcutil.LookupSRV("tcp", host)

		if err == nil && srvRecord != nil {
			resolvedHost = strings.Trim(srvRecord.Target, ".")
		}
	}

	// Resolve the connection hostname to an IP address
	{
		addr, err := net.ResolveIPAddr("ip", resolvedHost)

		if err == nil && addr != nil {
			ipAddress = PointerOf(addr.IP.String())
		}
	}

	statusContext, statusCancel := context.WithTimeout(context.Background(), opts.Timeout)
	legacyContext, legacyCancel := context.WithTimeout(context.Background(), opts.Timeout)
	queryContext, queryCancel := context.WithTimeout(context.Background(), opts.Timeout)

	defer statusCancel()
	defer legacyCancel()
	defer queryCancel()

	// Retrieve the post-netty rewrite Java Edition status (Minecraft 1.8+)
	{
		go func() {
			statusResult, _ = mcutil.Status(statusContext, host, port, options.JavaStatus{
				EnableSRV:       true,
				Timeout:         opts.Timeout - time.Millisecond*100,
				ProtocolVersion: 47,
				Ping:            false,
			})

			wg.Done()

			legacyCancel()

			if opts.Query && queryResult == nil {
				time.Sleep(time.Millisecond * 250)

				if queryResult == nil {
					queryCancel()
				}
			}
		}()
	}

	// Retrieve the pre-netty rewrite Java Edition status (Minecraft 1.7 and below)
	{
		go func() {
			legacyStatusResult, _ = mcutil.StatusLegacy(legacyContext, host, port, options.JavaStatusLegacy{
				EnableSRV:       true,
				Timeout:         opts.Timeout - time.Millisecond*100,
				ProtocolVersion: 47,
			})

			wg.Done()

			time.Sleep(time.Millisecond * 250)

			if queryResult == nil {
				queryCancel()
			}
		}()
	}

	// Retrieve the query information (if it is available)
	if opts.Query {
		go func() {
			queryResult, _ = mcutil.FullQuery(queryContext, host, port, options.Query{
				Timeout: opts.Timeout - time.Millisecond*100,
			})

			wg.Done()
		}()
	}

	wg.Wait()

	return BuildJavaResponse(host, port, statusResult, legacyStatusResult, queryResult, srvRecord, ipAddress)
}

// FetchBedrockStatus fetches a fresh status of a Bedrock Edition server.
func FetchBedrockStatus(host string, port uint16, opts *StatusOptions) BedrockStatusResponse {
	var (
		ipAddress *string
		status    *response.BedrockStatus
	)

	// Resolve the connection hostname to an IP address
	{
		ipAddr, err := net.ResolveIPAddr("ip", host)

		if err == nil && ipAddr != nil {
			ipAddress = PointerOf(ipAddr.IP.String())
		}
	}

	// Retrieve the Bedrock Edition status
	{
		ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)

		defer cancel()

		status, _ = mcutil.StatusBedrock(ctx, host, port)
	}

	return BuildBedrockResponse(host, port, status, ipAddress)
}

// BuildJavaResponse builds the response data from the status and query information.
func BuildJavaResponse(host string, port uint16, status *response.JavaStatus, legacyStatus *response.JavaStatusLegacy, query *response.FullQuery, srvRecord *net.SRV, ipAddress *string) (result JavaStatusResponse) {
	result = JavaStatusResponse{
		BaseStatus: BaseStatus{
			Online:      false,
			Host:        host,
			Port:        port,
			IPAddress:   ipAddress,
			EULABlocked: IsBlockedAddress(host),
			RetrievedAt: time.Now().UnixMilli(),
			ExpiresAt:   time.Now().Add(config.Cache.JavaStatusDuration).UnixMilli(),
		},
		JavaStatus: nil,
	}

	// Status
	if status != nil {
		result.Online = true

		result.JavaStatus = &JavaStatus{
			Version: &JavaVersion{
				NameRaw:   status.Version.NameRaw,
				NameClean: status.Version.NameClean,
				NameHTML:  status.Version.NameHTML,
				Protocol:  status.Version.Protocol,
			},
			Players: JavaPlayers{
				Online: status.Players.Online,
				Max:    status.Players.Max,
				List:   make([]Player, 0),
			},
			MOTD: MOTD{
				Raw:   status.MOTD.Raw,
				Clean: status.MOTD.Clean,
				HTML:  status.MOTD.HTML,
			},
			Icon:    nil,
			Mods:    make([]Mod, 0),
			Plugins: make([]Plugin, 0),
		}

		if status.Players.Sample != nil {
			for _, player := range status.Players.Sample {
				result.Players.List = append(result.Players.List, Player{
					UUID:      player.ID,
					NameRaw:   player.NameRaw,
					NameClean: player.NameClean,
					NameHTML:  player.NameHTML,
				})
			}
		}

		if status.Favicon != nil && len(*status.Favicon) > 0 {
			result.Icon = status.Favicon
		}

		if status.ModInfo != nil {
			for _, mod := range status.ModInfo.Mods {
				result.Mods = append(result.Mods, Mod{
					Name:    mod.ID,
					Version: mod.Version,
				})
			}
		}
	} else if legacyStatus != nil {
		result.Online = true

		result.JavaStatus = &JavaStatus{
			Version: nil,
			Players: JavaPlayers{
				Online: &legacyStatus.Players.Online,
				Max:    &legacyStatus.Players.Max,
				List:   make([]Player, 0),
			},
			MOTD: MOTD{
				Raw:   legacyStatus.MOTD.Raw,
				Clean: legacyStatus.MOTD.Clean,
				HTML:  legacyStatus.MOTD.HTML,
			},
			Icon:    nil,
			Mods:    make([]Mod, 0),
			Plugins: make([]Plugin, 0),
		}

		if legacyStatus.Version != nil {
			result.Version = &JavaVersion{
				NameRaw:   legacyStatus.Version.NameRaw,
				NameClean: legacyStatus.Version.NameClean,
				NameHTML:  legacyStatus.Version.NameHTML,
				Protocol:  legacyStatus.Version.Protocol,
			}
		}
	}

	// Query
	if query != nil {
		result.Online = true

		if result.JavaStatus == nil {
			result.JavaStatus = &JavaStatus{
				Players: JavaPlayers{
					List: make([]Player, 0),
				},
				Mods:    make([]Mod, 0),
				Plugins: make([]Plugin, 0),
			}

			if motd, ok := query.Data["hostname"]; ok {
				if parsedMOTD, err := formatting.Parse(motd); err == nil {
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

			if version, ok := query.Data["version"]; ok {
				parsedValue, err := formatting.Parse(version)

				if err == nil {
					result.Version = &JavaVersion{
						NameRaw:   parsedValue.Raw,
						NameClean: parsedValue.Clean,
						NameHTML:  parsedValue.HTML,
						Protocol:  0,
					}
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

			parsedName, err := formatting.Parse(username)

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

	if srvRecord != nil {
		result.SRVRecord = &SRVRecord{
			Host: strings.Trim(srvRecord.Target, "."),
			Port: srvRecord.Port,
		}
	}

	return
}

// BuildBedrockResponse builds the response data from the status information.
func BuildBedrockResponse(host string, port uint16, status *response.BedrockStatus, ipAddress *string) (result BedrockStatusResponse) {
	result = BedrockStatusResponse{
		BaseStatus: BaseStatus{
			Online:      false,
			Host:        host,
			Port:        port,
			IPAddress:   ipAddress,
			EULABlocked: IsBlockedAddress(host),
			RetrievedAt: time.Now().UnixMilli(),
			ExpiresAt:   time.Now().Add(config.Cache.BedrockStatusDuration).UnixMilli(),
		},
		BedrockStatus: nil,
	}

	if status != nil {
		result.Online = true

		result.BedrockStatus = &BedrockStatus{
			Version:  nil,
			Players:  nil,
			MOTD:     nil,
			Gamemode: status.Gamemode,
			ServerID: status.ServerID,
			Edition:  status.Edition,
		}

		if status.Version != nil {
			if result.Version == nil {
				result.Version = &BedrockVersion{
					Name:     nil,
					Protocol: nil,
				}
			}

			result.Version.Name = status.Version
		}

		if status.ProtocolVersion != nil {
			if result.Version == nil {
				result.Version = &BedrockVersion{
					Name:     nil,
					Protocol: nil,
				}
			}

			result.Version.Protocol = status.ProtocolVersion
		}

		if status.OnlinePlayers != nil {
			if result.Players == nil {
				result.Players = &BedrockPlayers{
					Online: nil,
					Max:    nil,
				}
			}

			result.Players.Online = status.OnlinePlayers
		}

		if status.MaxPlayers != nil {
			if result.Players == nil {
				result.Players = &BedrockPlayers{
					Online: nil,
					Max:    nil,
				}
			}

			result.Players.Max = status.MaxPlayers
		}

		if status.MOTD != nil {
			result.MOTD = &MOTD{
				Raw:   status.MOTD.Raw,
				Clean: status.MOTD.Clean,
				HTML:  status.MOTD.HTML,
			}
		}
	}

	return
}
