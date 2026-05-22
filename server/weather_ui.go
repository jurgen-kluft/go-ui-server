package remote_ui_server

import (
	fontpack "github.com/jurgen-kluft/go-gx2/fontpak"
	spritepack "github.com/jurgen-kluft/go-gx2/spritepak"
)

type WeatherUI struct {
	clientInfo ClientInfo
	spritePack *spritepack.SpritePack
	fontPack   *fontpack.FontPack
}

func NewWeatherUserInterface(spritePack *spritepack.SpritePack, fontPack *fontpack.FontPack) *WeatherUI {
	return &WeatherUI{
		spritePack: spritePack,
		fontPack:   fontPack,
	}
}

func (ui *WeatherUI) OnClientInfo(info ClientInfo) {
	ui.clientInfo = info
}

func (ui *WeatherUI) OnTouch(x, y int) {
	// Handle touch event at coordinates (x, y)
}

func (ui *WeatherUI) OnSwipe(direction, distance int) {
	// Handle swipe event with given direction and distance
}

func (ui *WeatherUI) OnButton(buttonId, state int) {
	// Handle button event with given button ID and state
}

func (ui *WeatherUI) OnRotary(rotation int) {
	// Handle rotary event with given rotation value
}

func (ui *WeatherUI) Render() (prevFrame *FrameBuffer, currFrame *FrameBuffer) {
	// Render the UI and return the previous and current frame data
	return nil, nil
}
