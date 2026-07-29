package remote_ui_server

import (
	"fmt"
	"net"
)

type AssetServer struct {
	config *Configuration
}

func NewAssetServer(config *Configuration) *AssetServer {
	server := &AssetServer{
		config: config,
	}

	return server
}

func (s *AssetServer) Start() error {

	// Load all sprite packs and font packs into memory, so they can be shared across user interfaces.
	assetDb, err := BuildAssetDatabase(s.config.Assets)
	if err != nil {
		return fmt.Errorf("failed to build asset database: %v", err)
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
			fmt.Println("Error starting Asset Server:", err)
			return
		}
		defer listener.Close()

		fmt.Printf("Asset Server listening on port %d\n", s.config.Port)

		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("Error accepting connection:", err)
				continue
			}

			// Read the ClientInfo message from the connection, and parse it into a ClientInfo struct.
			go func() {
				defer conn.Close()

				// The first message from the client is the MacAddress, which is used to identify the client
				// and determine if it is allowed to connect.
				macAddressStr, err := readClientInfo(conn)
				if err != nil {
					fmt.Println("Error reading ClientInfo:", err)
					return
				}

				if client, ok := s.config.Clients[macAddressStr]; ok {
					// Launch the upload task in a new go-routine, so it can run concurrently and independently.
					scriptBytes := assetDb.GetScriptBytesFor(client.Name)
					assetDbBytes := assetDb.GetBytesFor(client.Name)
					go func() {
						ClientUpload(conn, scriptBytes, assetDbBytes)
					}()
				}
			}()
		}

	}()
	return nil
}

func readClientInfo(conn net.Conn) (string, error) {
	// Read 8 bytes from the connection, and parse it into the ClientInfo struct.
	buf := make([]byte, 8)
	_, err := conn.Read(buf)
	if err != nil {
		return "", err
	}

	macAddress := fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		buf[0], buf[1], buf[2], buf[3], buf[4], buf[5],
	)
	return macAddress, nil
}
