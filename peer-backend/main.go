package main

import (
	"fmt"
	"log"
	"os"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/wasif9/p2p-file-share/discovery"
	"github.com/wasif9/p2p-file-share/messaging"
	"github.com/wasif9/p2p-file-share/network"
	"github.com/wasif9/p2p-file-share/transfer"
)

func main() {
	// Create a new libp2p node (peer)
	node, err := libp2p.New()
	if err != nil {
		log.Fatal(err)
	}

	// Print the node's multiaddresses
	fmt.Println("Node ID:", node.ID().String())
	for _, addr := range node.Addrs() {
		fmt.Println("Listening on:", addr)
	}

	// Enable DHT for file lookup
	// dht.SetupDHT(node)  // Commented out for now

	// Enable mDNS-based peer discovery
	discovery.SetupDiscovery(node)

	// Handle incoming messages
	messaging.HandleMessages(node)

	// Handle incoming file requests
	transfer.HandleFileRequests(node)

	if len(os.Args) > 3 && os.Args[1] == "send-file" {
		targetAddr := os.Args[2]
		filePath := os.Args[3]

		// Connect to the target peer
		network.ConnectToPeer(node, targetAddr)

		// ⬇️ Add this delay to ensure connection is fully established
		time.Sleep(1 * time.Second)

		// Extract peer ID from multiaddr
		addr, _ := multiaddr.NewMultiaddr(targetAddr)
		peerInfo, _ := peer.AddrInfoFromP2pAddr(addr)
		fmt.Println("Attempting to send file to peer")

		// Send a file to the target peer
		transfer.SendFile(node, peerInfo.ID, filePath)
	}

	// Keep the node running
	select {}
}
