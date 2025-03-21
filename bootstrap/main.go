package main

import (
	"context"
	"fmt"
	"log"

	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
)

func main() {
	ctx := context.Background()

	// Create a Libp2p Host that listens on all available interfaces (server mode)
	node, err := libp2p.New()
	if err != nil {
		log.Fatal("Failed to create libp2p host:", err)
	}

	// Create the Kademlia DHT explicitly in SERVER MODE.
	kad, err := dht.New(ctx, node, dht.Mode(dht.ModeServer))
	if err != nil {
		log.Fatal("Failed to create DHT:", err)
	}

	// Bootstrap your DHT server to populate initial routing tables
	if err := kad.Bootstrap(ctx); err != nil {
		log.Fatal("Failed to bootstrap DHT:", err)
	}

	// Display your bootstrap node's multiaddrs clearly
	fmt.Println("🌟 Bootstrap Node running successfully! ✅")
	fmt.Println("Peer ID:", node.ID())
	fmt.Println("Bootstrap multiaddrs:")
	for _, addr := range node.Addrs() {
		fmt.Printf("%s/p2p/%s\n", addr, node.ID())
	}

	// Prevent the application from exiting (keep alive)
	select {}
}
