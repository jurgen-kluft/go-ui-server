package remote_ui_server

import (
	"encoding/json"
	"os"
)

type FontDescriptor struct {
	Id   int    `json:"Id"`
	Path string `json:"Path"`
}

type SpriteDescriptor struct {
	Id   int    `json:"Id"`
	Path string `json:"Path"`
}

type RemoteClientDescriptor struct {
	Name       string `json:"Name"`
	MacAddress string `json:"MacAddress"`
}

type Configuration struct {
	// The port on which the server listens for incoming connections.
	Port int `json:"Port"`

	FontPackCfgFile    string `json:"FontPackCfgFile"`
	SpritePackCfgFile  string `json:"SpritePackCfgFile"`
	PalettePackCfgFile string `json:"PalettePackCfgFile"`
	ScriptFile         string `json:"ScriptFile"`

	// List of remote clients that are allowed to connect, identified by their MAC address
	// and the user interface they should be served.
	AllowedClients []string `json:"AllowedClients"`
}

func LoadConfig(filepath string) (*Configuration, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	config := &Configuration{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}
