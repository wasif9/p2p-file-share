package main

import (
	"context"
	"log"
	"os"

	"encoding/base64"

	"github.com/joho/godotenv"
	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
)

func loadStaticIdentity(base64PrivKey string) (crypto.PrivKey, error) {
	privBytes, err := base64.StdEncoding.DecodeString(base64PrivKey)
	if err != nil {
		return nil, err
	}
	priv, err := crypto.UnmarshalPrivateKey(privBytes)
	return priv, err
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	ctx := context.Background()

	var node host.Host
	var nodeErr error

	// Create a Libp2p Host that listens on all available interfaces (server mode)
	if err := godotenv.Load(); err != nil {
		log.Println("Bootstrap node fail to load .env file, use random ID", err)

		node, nodeErr = libp2p.New(
			libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/8001"),
		)
	} else {
		base64PrivKey := os.Getenv("BOOTSTRAP_PRIVKEY")
		bootstrapKey, err := loadStaticIdentity(base64PrivKey)

		if err != nil {
			log.Println("Bootstrap node fail to decode the private key in .env file, use random ID", err)

			node, nodeErr = libp2p.New(
				libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/8001"),
			)
		} else {
			node, nodeErr = libp2p.New(
				libp2p.Identity(bootstrapKey),
				libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/8001"),
			)
		}
	}
	if nodeErr != nil {
		log.Fatal("Failed to create libp2p host:", nodeErr)
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
	log.Println("🌟 Bootstrap Node running successfully! ✅")
	log.Println("Peer ID:", node.ID())
	log.Println("Bootstrap multiaddrs:")
	for _, addr := range node.Addrs() {
		log.Printf("%s/p2p/%s\n", addr, node.ID())
	}

	// Prevent the application from exiting (keep alive)
	select {}
}
