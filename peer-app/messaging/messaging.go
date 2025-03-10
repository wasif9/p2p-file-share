package messaging

import (
	"context"
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

// Handle incoming messages
func HandleMessages(node host.Host) {
	node.SetStreamHandler("/chat/1.0.0", func(s network.Stream) {
		defer s.Close() // Close stream when done

		buf := make([]byte, 256)
		n, err := s.Read(buf)
		if err != nil {
			log.Println("Failed to read:", err)
			return
		}
		fmt.Println("Received message:", string(buf[:n]))
	})
}

// Send a message to a peer
func SendMessage(node host.Host, peerID peer.ID, message string) {
	s, err := node.NewStream(context.Background(), peerID, "/chat/1.0.0")
	if err != nil {
		log.Fatal("Failed to open stream:", err)
	}
	defer s.Close() // Close stream after sending

	_, err = s.Write([]byte(message))
	if err != nil {
		log.Fatal("Failed to send message:", err)
	}
	fmt.Println("Message sent:", message)
}
