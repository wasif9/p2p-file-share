package network

import (
	"context"
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// Connect to a peer using multiaddress
func ConnectToPeer(node host.Host, targetAddr string) {
	addr, err := multiaddr.NewMultiaddr(targetAddr)
	if err != nil {
		log.Fatal("Invalid address:", err)
	}

	peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		log.Fatal("Failed to parse peer address:", err)
	}

	if err := node.Connect(context.Background(), *peerInfo); err != nil {
		log.Fatal("Failed to connect:", err)
	}

	fmt.Println("Connected to:", peerInfo.ID.String())
}
