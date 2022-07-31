package main

import "github.com/PassTheMayo/mcstatus/v4"

type StatusResponse[T JavaStatus | BedrockStatus] struct {
	Online      bool   `json:"online"`
	Host        string `json:"host"`
	Port        uint16 `json:"port"`
	EULABlocked bool   `json:"eula_blocked"`
	Response    *T     `json:"response"`
}

type JavaStatus struct {
	Version   *Version   `json:"version"`
	Players   Players    `json:"players"`
	MOTD      MOTD       `json:"motd"`
	Favicon   *string    `json:"favicon"`
	ModInfo   *ModInfo   `json:"mod_info"`
	SRVRecord *SRVRecord `json:"srv_record"`
}

type Players struct {
	Online int            `json:"online"`
	Max    int            `json:"max"`
	Sample []SamplePlayer `json:"sample"`
}

type SamplePlayer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Clean string `json:"clean"`
	HTML  string `json:"html"`
}

type ModInfo struct {
	Type string `json:"type"`
	Mods []Mod  `json:"mods"`
}

type Mod struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

type BedrockStatus struct {
	ServerGUID      int64      `json:"server_guid"`
	Edition         *string    `json:"edition"`
	MOTD            *MOTD      `json:"motd"`
	ProtocolVersion *int64     `json:"protocol_version"`
	Version         *string    `json:"version"`
	OnlinePlayers   *int64     `json:"online_players"`
	MaxPlayers      *int64     `json:"max_players"`
	ServerID        *string    `json:"server_id"`
	Gamemode        *string    `json:"gamemode"`
	GamemodeID      *int64     `json:"gamemode_id"`
	PortIPv4        *uint16    `json:"port_ipv4"`
	PortIPv6        *uint16    `json:"port_ipv6"`
	SRVRecord       *SRVRecord `json:"srv_record"`
}

type MOTD struct {
	Raw   string `json:"raw"`
	Clean string `json:"clean"`
	HTML  string `json:"html"`
}

type Version struct {
	Name     string `json:"name"`
	Protocol int    `json:"protocol"`
}

type SRVRecord struct {
	Host string `json:"host"`
	Port uint16 `json:"port"`
}

func GetJavaStatus(host string, port uint16) (resp StatusResponse[JavaStatus]) {
	status, err := mcstatus.Status(host, port)

	if err != nil {
		statusLegacy, err := mcstatus.StatusLegacy(host, port)

		if err != nil {
			resp = StatusResponse[JavaStatus]{
				Online:      false,
				Host:        host,
				Port:        port,
				EULABlocked: IsBlockedAddress(host),
				Response:    nil,
			}

			return
		}

		resp = StatusResponse[JavaStatus]{
			Online:      true,
			Host:        host,
			Port:        port,
			EULABlocked: IsBlockedAddress(host),
			Response: &JavaStatus{
				Version: nil,
				Players: Players{
					Online: statusLegacy.Players.Online,
					Max:    statusLegacy.Players.Max,
					Sample: make([]SamplePlayer, 0),
				},
				MOTD: MOTD{
					Raw:   statusLegacy.MOTD.Raw,
					Clean: statusLegacy.MOTD.Clean,
					HTML:  statusLegacy.MOTD.HTML,
				},
				Favicon:   nil,
				ModInfo:   nil,
				SRVRecord: nil,
			},
		}

		if statusLegacy.Version != nil {
			resp.Response.Version = &Version{
				Name:     statusLegacy.Version.Name,
				Protocol: statusLegacy.Version.Protocol,
			}
		}

		if statusLegacy.SRVResult != nil {
			resp.Response.SRVRecord = &SRVRecord{
				Host: statusLegacy.SRVResult.Host,
				Port: statusLegacy.SRVResult.Port,
			}
		}

		return
	}

	samplePlayers := make([]SamplePlayer, 0)

	for _, player := range status.Players.Sample {
		samplePlayers = append(samplePlayers, SamplePlayer{
			ID:    player.ID,
			Name:  player.Name,
			Clean: player.Clean,
			HTML:  player.HTML,
		})
	}

	resp = StatusResponse[JavaStatus]{
		Online:      true,
		Host:        host,
		Port:        port,
		EULABlocked: IsBlockedAddress(host),
		Response: &JavaStatus{
			Version: &Version{
				Name:     status.Version.Name,
				Protocol: status.Version.Protocol,
			},
			Players: Players{
				Online: status.Players.Online,
				Max:    status.Players.Max,
				Sample: samplePlayers,
			},
			MOTD: MOTD{
				Raw:   status.MOTD.Raw,
				Clean: status.MOTD.Clean,
				HTML:  status.MOTD.HTML,
			},
			Favicon:   status.Favicon,
			ModInfo:   nil,
			SRVRecord: nil,
		},
	}

	if status.ModInfo != nil {
		mods := make([]Mod, 0)

		for _, mod := range status.ModInfo.Mods {
			mods = append(mods, Mod{
				ID:      mod.ID,
				Version: mod.Version,
			})
		}

		resp.Response.ModInfo = &ModInfo{
			Type: status.ModInfo.Type,
			Mods: mods,
		}
	}

	if status.SRVResult != nil {
		resp.Response.SRVRecord = &SRVRecord{
			Host: status.SRVResult.Host,
			Port: status.SRVResult.Port,
		}
	}

	return
}

func GetBedrockStatus(host string, port uint16) (resp StatusResponse[BedrockStatus]) {
	status, err := mcstatus.StatusBedrock(host, port)

	if err != nil {
		resp = StatusResponse[BedrockStatus]{
			Online:      false,
			Host:        host,
			Port:        port,
			EULABlocked: IsBlockedAddress(host),
			Response:    nil,
		}

		return
	}

	resp = StatusResponse[BedrockStatus]{
		Online:      true,
		Host:        host,
		Port:        port,
		EULABlocked: IsBlockedAddress(host),
		Response: &BedrockStatus{
			ServerGUID:      status.ServerGUID,
			Edition:         status.Edition,
			MOTD:            nil,
			ProtocolVersion: status.ProtocolVersion,
			Version:         status.Version,
			OnlinePlayers:   status.OnlinePlayers,
			MaxPlayers:      status.MaxPlayers,
			ServerID:        status.ServerID,
			Gamemode:        status.Gamemode,
			GamemodeID:      status.GamemodeID,
			PortIPv4:        status.PortIPv4,
			PortIPv6:        status.PortIPv6,
			SRVRecord:       nil,
		},
	}

	if status.MOTD != nil {
		resp.Response.MOTD = &MOTD{
			Raw:   status.MOTD.Raw,
			Clean: status.MOTD.Clean,
			HTML:  status.MOTD.HTML,
		}
	}

	if status.SRVResult != nil {
		resp.Response.SRVRecord = &SRVRecord{
			Host: status.SRVResult.Host,
			Port: status.SRVResult.Port,
		}
	}

	return
}
