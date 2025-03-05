package main

import (
	"fmt"
	"log"
	"os"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/wasif9/p2p-file-share/discovery" // Import discovery package
	"github.com/wasif9/p2p-file-share/messaging" // Import messaging package
	"github.com/wasif9/p2p-file-share/network"   // Import network package
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

	// Enable mDNS-based peer discovery
	discovery.SetupDiscovery(node)

	// Handle incoming messages
	messaging.HandleMessages(node)

	if len(os.Args) > 2 {
		targetAddr := os.Args[1]
		message := os.Args[2]

		// Connect to the target peer
		network.ConnectToPeer(node, targetAddr)

		// Extract peer ID from multiaddr
		addr, _ := multiaddr.NewMultiaddr(targetAddr)
		peerInfo, _ := peer.AddrInfoFromP2pAddr(addr)

		// Send a message to the target peer
		messaging.SendMessage(node, peerInfo.ID, message)
	}

	// Keep the node running
	select {}
}
