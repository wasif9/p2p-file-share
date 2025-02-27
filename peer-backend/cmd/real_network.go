package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RealNetworkAdapter sends messages via HTTP
type RealNetworkAdapter struct {
	Nodes map[string]string // NodeID -> Ngrok Public URL
}

// SendMessage sends a message to another node, retrying if unavailable
func (r *RealNetworkAdapter) SendMessage(targetID string, message string) error {
	targetAddress, exists := r.Nodes[targetID]
	if !exists {
		return fmt.Errorf("target node not found: %s", targetID)
	}

	url := targetAddress + "/rpc/message"
	reqBody := bytes.NewBuffer([]byte(message))

	// Retry mechanism: Try 5 times with a delay
	for i := 1; i <= 5; i++ {
		resp, err := http.Post(url, "text/plain", reqBody)
		if err != nil {
			fmt.Printf("Attempt %d: Failed to send message to %s, retrying...\n", i, targetID)
			time.Sleep(3 * time.Second) // Wait before retrying
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Response from %s: %s\n", targetID, string(body))
		return nil
	}

	// If all retries fail, return error
	return fmt.Errorf("failed to send message to %s after multiple attempts", targetID)
}
