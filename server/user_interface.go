package remote_ui_server

const (
	MessageTypeClientInfo MessageType = 0xC001
)

type ClientInfo struct {
	MessageType  uint16  // should be MessageTypeClientInfo
	MessageLen   uint16  // 16 (fixed size of the ClientInfo struct)
	DeviceFormat uint16  //
	ScreenWidth  uint16  // in pixels
	ScreenHeight uint16  // in pixels
	MacAddress   [6]byte // MAC address of the client device
}

type RemoteUserInterface interface {
	OnClientInfo(info ClientInfo)

	OnTouch(x, y int)
	OnSwipe(direction int, distance int)
	OnButton(buttonID int, state int)
	OnRotary(rotation int)

	Render() (previous *FrameBuffer, current *FrameBuffer)
}
