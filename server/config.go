package remote_ui_server

import (
	"encoding/json"
	"os"
)

type ClientDescriptor struct {
	Name       string `json:"Name"`
	MacAddress string `json:"MacAddress"`
	Assets     string `json:"Assets"`
}

type AssetDescriptor struct {
	Name               string `json:"Name"`
	FontPackCfgFile    string `json:"FontPackCfgFile"`
	SpritePackCfgFile  string `json:"SpritePackCfgFile"`
	PalettePackCfgFile string `json:"PalettePackCfgFile"`
	ScriptFile         string `json:"ScriptFile"`
}

type Configuration struct {
	Port     int                         `json:"Port"`
	CacheDir string                      `json:"CacheDir"`
	Assets   map[string]AssetDescriptor  `json:"Assets"`
	Clients  map[string]ClientDescriptor `json:"Clients"`
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
