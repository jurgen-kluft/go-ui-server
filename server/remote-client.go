package remote_ui_server

import (
	"encoding/binary"
	"fmt"
	"net"
)

type RemoteClient struct {
	tcpConn net.Conn
}

// Message Types
type MessageType uint16

const (
	MessageTypeAsset  MessageType = 0xC002
	MessageTypeScript MessageType = 0xC003
)

func ClientUpload(conn net.Conn, scriptBytes []byte, assetDbBytes []byte) error {

	blockSize := 8192 // 8 KB block size
	blockIndex := 0
	blockCount := (len(assetDbBytes) + blockSize - 1) / blockSize // Calculate the number of blocks needed

	sendBuffer := make([]byte, 16384)

	for {
		// Send the asset database to the client, block by block.
		// The format of the message is:
		// - 2 bytes: message type (0xC002 for asset database)
		// - 2 bytes: length of the asset database in bytes
		// - 2 bytes: number of blocks (each block is 8 KB)
		// - 2 bytes: block index (0-based)
		// - 8 KB: block data (last block may be smaller than 8 KB)

		if blockIndex < blockCount {
			// Calculate the start and end indices for the current block
			start := blockIndex * blockSize
			end := start + blockSize
			if end > len(assetDbBytes) {
				end = len(assetDbBytes)
			}

			// Prepare the message header
			const headerSize = 8
			messageType := MessageTypeAsset
			messageLength := uint16(end - start)
			numBlocks := uint16(blockCount)
			blockIdx := uint16(blockIndex)

			// Create a buffer for the message
			binary.BigEndian.PutUint16(sendBuffer[0:2], uint16(messageType))
			binary.BigEndian.PutUint16(sendBuffer[2:4], messageLength)
			binary.BigEndian.PutUint16(sendBuffer[4:6], numBlocks)
			binary.BigEndian.PutUint16(sendBuffer[6:8], blockIdx)

			// Copy the block data into the buffer
			copy(sendBuffer[headerSize:], assetDbBytes[start:end])

			// Send the message to the client
			_, err := conn.Write(sendBuffer[:headerSize+(end-start)])
			if err != nil {
				return fmt.Errorf("error sending asset database block %d: %v", blockIndex, err)
			}

			// Move to the next block
			blockIndex++
		} else {
			// All blocks have been sent, exit the loop
			break
		}
	}

	// After sending the asset database, send the script file to the client.
	// The format of the message is:
	// - 2 bytes: message type (0xC003 for script file)
	// - 2 bytes: length of the script file in bytes
	// - N bytes: script file data

	messageType := MessageTypeScript
	messageLength := uint16(len(scriptBytes))

	// Create a buffer for the message
	const headerSize = 4
	binary.BigEndian.PutUint16(sendBuffer[0:2], uint16(messageType))
	binary.BigEndian.PutUint16(sendBuffer[2:4], messageLength)

	// Copy the script file data into the buffer
	copy(sendBuffer[headerSize:], scriptBytes)

	// Send the message to the client
	_, err := conn.Write(sendBuffer[:headerSize+len(scriptBytes)])
	if err != nil {
		return fmt.Errorf("error sending script file: %v", err)
	}

	return nil
}
