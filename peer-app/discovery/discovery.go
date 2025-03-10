package discovery

import (
	"context"
	"log"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	mdns "github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

// Notifee implements the mdns.Notifee interface to handle peer discovery events.
type notifee struct {
	node host.Host
}

// HandlePeerFound is called when a new peer is discovered via mDNS.
func (n *notifee) HandlePeerFound(pi peer.AddrInfo) {
	log.Println("Discovered new peer:", pi.ID.String())

	// Attempt to connect to the discovered peer
	if err := n.node.Connect(context.Background(), pi); err != nil {
		log.Println("Failed to connect to discovered peer:", err)
	} else {
		log.Println("Successfully connected to discovered peer:", pi.ID.String())
	}
}

// SetupDiscovery initializes mDNS service for peer discovery.
func SetupDiscovery(node host.Host) {
	// Create a new mDNS service with a service name (p2p-demo)
	service := mdns.NewMdnsService(node, "_p2p._udp", &notifee{node: node})
	if service == nil {
		log.Fatal("Failed to create mDNS service")
	}

	// ⬇️ Explicitly start the mDNS service
	if err := service.Start(); err != nil {
		log.Fatal("Failed to start mDNS service:", err)
	}
	log.Println("mDNS service started successfully")

	// Keep the mDNS service running
	//defer service.Close()
}
