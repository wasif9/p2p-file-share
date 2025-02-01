package main

import (
	"fmt"
)

// FakeNetworkAdapter simulates P2P communication in memory
type FakeNetworkAdapter struct {
	Nodes map[string]*Node // Map of nodeID -> Node instance
}

// SendMessage simulates sending a message to another node
func (f *FakeNetworkAdapter) SendMessage(targetID string, message string) error {
	targetNode, ok := f.Nodes[targetID]
	if !ok {
		return fmt.Errorf("target node not found: %s", targetID)
	}

	// Call the target node's message handler directly
	targetNode.HandleMessage(message)
	return nil
}
