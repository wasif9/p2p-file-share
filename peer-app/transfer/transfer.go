package transfer

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

// Protocol ID for file transfer
const FileTransferProtocol = "/file-transfer/1.0.0"

// SendFile sends a file to a peer using a libp2p stream
func SendFile(node host.Host, peerID peer.ID, filePath string) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Failed to open file:", err)
	}
	defer file.Close()

	// Create a new stream to the peer
	s, err := node.NewStream(context.Background(), peerID, FileTransferProtocol)
	if err != nil {
		log.Fatal("Failed to create stream:", err)
	}
	defer s.Close()

	// Send the file
	log.Println("Sending file:", filePath)
	_, err = io.Copy(s, file)
	if err != nil {
		log.Fatal("Failed to send file:", err)
	}

	log.Println("File sent successfully!")
}

// HandleFileRequests listens for incoming file requests and sends files
func HandleFileRequests(node host.Host) {
	node.SetStreamHandler(FileTransferProtocol, func(s network.Stream) {
		defer s.Close()

		// Create or overwrite the received file
		file, err := os.Create("received_myfile.txt")
		if err != nil {
			log.Println("Failed to create file:", err)
			return
		}
		defer file.Close()

		// Receive the file and save it
		log.Println("Receiving file...")
		_, err = io.Copy(file, s)
		if err != nil {
			log.Println("Failed to receive file:", err)
			return
		}

		log.Println("File received successfully and saved as 'received_myfile.txt'")
	})
}
