package remote_ui_server

import (
	"fmt"
	"net"
	"os"
	"sync"

	fontpack "github.com/jurgen-kluft/go-gx2/fontpak"
	spritepack "github.com/jurgen-kluft/go-gx2/spritepak"
)

func UserInterfaceFactory(name string, sp *spritepack.SpritePack, fp *fontpack.FontPack) RemoteUserInterface {
	switch name {
	case "Weather UI":
		return NewWeatherUserInterface(sp, fp)
	default:
		return nil
	}
}

// The Remote UI server, which listens for incoming TCP connections from clients,
// and creates a new instance for each connection. Each instance runs in its own
// go-routine, and manages the lifecycle of the connection, including receiving
// messages from the client, processing them, and sending responses back.

// The Remote UI server when starting does the following:
// - Load configuration
// - Collect all sprite pack and font pack paths in a set so as to only load each
//   pack once, even if multiple user interfaces use the same pack.
// - Load all Sprite and Font packs into memory, and store them in a map.
// - Start TCP server and listen for incoming connections. For each new connection:
//   - Wait for ClientInfo message, which contains the name of the UI the client wants
//     to use, and the client's MAC address.
//   - Create a new UserInterface instance using the UserInterfaceFactory, passing in
//     the requested UI name and the loaded asset packs.
//   - Launch it on a new go-routine, so it can run concurrently and independently.

type RemoteClientInfo struct {
	IPAddress         string
	MACAddress        string
	ClientInfo        ClientInfo
	UserInterfaceName string
	UserInterface     RemoteUserInterface
}

type RemoteUserInterfaceServer struct {
	config *Configuration

	SpritePacks map[string]*spritepack.SpritePack
	FontPacks   map[string]*fontpack.FontPack

	// Client connection initiation is on a go-routine, so we need to make the access to the Clients
	// map thread-safe. We can use a sync.RWMutex to protect it, or we can use a sync.Map for simplicity.

	// When a client connects, we remember them fully, their IP address, MAC address, and
	// the UI they want to use. This allows us to deal with a client disconnecting and reconnecting,
	// and we can restore their session if they come back.
	ClientDescriptors map[string]RemoteClientDescriptor
	Clients           sync.Map
}

func (s *RemoteUserInterfaceServer) GetSpriteAndFontPackFor(uiName string) (*spritepack.SpritePack, *fontpack.FontPack) {
	// Find the UserInterfaceDescriptor for the given UI name, and return the corresponding SpritePack and FontPack from the server's maps.
	var uiDescriptor *UserInterfaceDescriptor
	for _, ui := range s.config.UserInterfaces {
		if ui.Name == uiName {
			uiDescriptor = &ui
			break
		}
	}

	if uiDescriptor == nil {
		return nil, nil
	}

	sp := s.SpritePacks[uiDescriptor.SpritePack]
	fp := s.FontPacks[uiDescriptor.FontPack]

	return sp, fp
}

func NewRemoteUserInterfaceServer(config *Configuration) *RemoteUserInterfaceServer {
	return &RemoteUserInterfaceServer{
		config:      config,
		SpritePacks: make(map[string]*spritepack.SpritePack),
		FontPacks:   make(map[string]*fontpack.FontPack),
		Clients:     sync.Map{},
	}
}

func (s *RemoteUserInterfaceServer) Start() error {
	// Build the ClientDescriptors map from the RemoteClients list in the configuration, so we can easily
	// look up a client by their MAC address when they connect.
	s.ClientDescriptors = make(map[string]RemoteClientDescriptor)
	for _, cd := range s.config.RemoteClients {
		s.ClientDescriptors[cd.MacAddress] = cd
	}

	// Load all sprite packs and font packs into memory, so they can be shared across user interfaces.
	for _, spd := range s.config.SpritePacks {
		f, err := os.Open(spd.Path)
		if err != nil {
			return err
		}
		defer f.Close()
		sp, err := spritepack.ReadPack(f)
		if err != nil {
			return err
		}
		s.SpritePacks[spd.Name] = sp
	}

	for _, fpd := range s.config.FontPacks {
		f, err := os.Open(fpd.Path)
		if err != nil {
			return err
		}
		defer f.Close()
		fp, err := fontpack.ReadPack(f)
		if err != nil {
			return err
		}
		s.FontPacks[fpd.Name] = fp
	}

	// Start TCP server and listen for incoming connections. For each new connection:
	// - Wait for ClientInfo message, which contains the name of the UI the client wants
	//   to use, and the client's MAC address.
	// - Create a new UserInterface instance using the UserInterfaceFactory, passing in
	//   the requested UI name and the loaded asset packs.
	// - Launch it on a new go-routine, so it can run concurrently and independently.
	go func() {

		// Listen for incoming TCP connections on the configured port, and handle them in a loop.
		// For each new connection, read the ClientInfo message, create a new UserInterface instance,
		// and launch it in a new go-routine.

		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
		if err != nil {
			fmt.Println("Error starting TCP server:", err)
			return
		}
		defer listener.Close()

		fmt.Printf("Remote UI Server listening on port %d\n", s.config.Port)

		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("Error accepting connection:", err)
				continue
			}

			// type ClientInfo struct {
			// 	MessageType   uint16   // should be MessageTypeClientInfo
			// 	MessageLen    uint16   // 32 (fixed size of the ClientInfo struct)
			// 	DeviceType    uint16   // 0=unknown, 1=phone, 2=tablet, 3=desktop
			// 	ScreenWidth   uint16   // in pixels
			// 	ScreenHeight  uint16   // in pixels
			// 	MacAddress    [6]byte  // MAC address of the client device
			// 	UserInterface [16]byte // Name of the requested user interface, null-terminated string
			// }
			// Read the ClientInfo message from the connection, and parse it into a ClientInfo struct.
			go func() {
				defer conn.Close()

				// Read the ClientInfo message from the connection, and parse it into a ClientInfo struct.
				var clientInfo ClientInfo
				err := readClientInfo(conn, &clientInfo)
				if err != nil {
					fmt.Println("Error reading ClientInfo:", err)
					return
				}

				// See if we already have a client with this MacAddress, if not create a new RemoteClientInfo for it.
				macAddressStr := fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
					clientInfo.MacAddress[0], clientInfo.MacAddress[1], clientInfo.MacAddress[2],
					clientInfo.MacAddress[3], clientInfo.MacAddress[4], clientInfo.MacAddress[5],
				)

				remoteClientInfoAny, exists := s.Clients.Load(macAddressStr)
				var remoteClientInfo *RemoteClientInfo
				if !exists {
					remoteClientInfo = &RemoteClientInfo{MACAddress: macAddressStr, ClientInfo: clientInfo}
					s.Clients.Store(macAddressStr, remoteClientInfo)
				} else {
					remoteClientInfo = remoteClientInfoAny.(*RemoteClientInfo)
				}

				// Create a new UserInterface instance using the UserInterfaceFactory, passing in
				// the requested UI name and the loaded asset packs.
				if remoteClientInfo.UserInterface == nil {
					if remoteClientDescriptor, ok := s.ClientDescriptors[remoteClientInfo.MACAddress]; ok {
						remoteClientInfo.UserInterfaceName = remoteClientDescriptor.UserInterface
						spritePak, fontPak := s.GetSpriteAndFontPackFor(remoteClientInfo.UserInterfaceName)
						ui := UserInterfaceFactory(remoteClientInfo.UserInterfaceName, spritePak, fontPak)
						remoteClientInfo.UserInterface = ui
					}
				}

				if remoteClientInfo.UserInterface != nil {
					// Launch the UserInterface instance in a new go-routine, so it can run concurrently and independently.
					go func() {
						remoteClient := &RemoteClient{
							tcpConn:       conn,
							userInterface: remoteClientInfo.UserInterface,
						}
						remoteClient.Run()
					}()
				} else {
					fmt.Printf("Unknown user interface '%s' requested for remote client with Mac %v\n", remoteClientInfo.UserInterfaceName, remoteClientInfo.ClientInfo.MacAddress)
				}

			}()
		}

	}()
	return nil
}

func readClientInfo(conn net.Conn, clientInfo *ClientInfo) error {
	// Read 16 bytes from the connection, and parse it into the ClientInfo struct.
	buf := make([]byte, 16)
	_, err := conn.Read(buf)
	if err != nil {
		return err
	}

	clientInfo.MessageType = uint16(buf[0])<<8 | uint16(buf[1])
	clientInfo.MessageLen = uint16(buf[2])<<8 | uint16(buf[3])
	clientInfo.DeviceFormat = uint16(buf[4])<<8 | uint16(buf[5])
	clientInfo.ScreenWidth = uint16(buf[6])<<8 | uint16(buf[7])
	clientInfo.ScreenHeight = uint16(buf[8])<<8 | uint16(buf[9])
	copy(clientInfo.MacAddress[:], buf[10:16])

	return nil
}
