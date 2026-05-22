package remote_ui_server

import (
	"encoding/json"
	"os"
)

// see ui_server/config.json

type FontPackDescriptor struct {
	Name string `json:"Name"`
	Path string `json:"Path"`
}

type SpritePackDescriptor struct {
	Name string `json:"Name"`
	Path string `json:"Path"`
}

type UserInterfaceDescriptor struct {
	Name       string `json:"Name"`
	FontPack   string `json:"FontPack"`
	SpritePack string `json:"SpritePack"`
}

type RemoteClientDescriptor struct {
	Name          string `json:"Name"`
	MacAddress    string `json:"MacAddress"`
	UserInterface string `json:"UserInterface"`
}

type Configuration struct {
	// The port on which the server listens for incoming connections.
	Port int `json:"Port"`

	// Available font packs that can be referenced by user interfaces.
	FontPacks []FontPackDescriptor `json:"FontPacks"`

	// Available sprite packs that can be referenced by user interfaces.
	SpritePacks []SpritePackDescriptor `json:"SpritePacks"`

	// Available user interfaces and the pack names they reference.
	UserInterfaces []UserInterfaceDescriptor `json:"UserInterfaces"`

	// List of remote clients that are allowed to connect, identified by their MAC address
	// and the user interface they should be served.
	RemoteClients []RemoteClientDescriptor `json:"RemoteClients"`
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
