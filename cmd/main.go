package main

import (
	"flag"
	"fmt"
	"sync"
	"time"
)

func main() {
	// Use a flag to determine whether to start Node A or Node B
	nodeID := flag.String("id", "", "Specify node ID (A or B)")
	flag.Parse()

	if *nodeID == "" {
		fmt.Println("Usage: go run main.go -id=<A|B>")
		return
	}

	var wg sync.WaitGroup
	wg.Add(1) // Only one node runs per instance

	network := &RealNetworkAdapter{Nodes: make(map[string]string)}

	// Manually set Ngrok public URLs for each node
	nodeAURL := "http://localhost:8081" // Replace with actual Ngrok URL
	nodeBURL := "http://localhost:8082" // Replace with actual Ngrok URL

	// Register nodes with their Ngrok URLs
	network.Nodes["A"] = nodeAURL
	network.Nodes["B"] = nodeBURL

	if *nodeID == "A" {
		// Start Node A
		nodeA := &Node{ID: "A", Address: "localhost:8081", Adapter: network}

		go func() {
			defer wg.Done()
			nodeA.StartServer()
		}()

		// Allow server to start
		time.Sleep(1 * time.Second)

		// Send message to B
		fmt.Println("Node A sending message to B...")
		nodeA.Adapter.SendMessage("B", "Hello from Node A!")

	} else if *nodeID == "B" {
		// Start Node B
		nodeB := &Node{ID: "B", Address: "localhost:8082", Adapter: network}

		go func() {
			defer wg.Done()
			nodeB.StartServer()
		}()

		// Allow server to start
		time.Sleep(1 * time.Second)

		// Send message to A
		fmt.Println("Node B sending message to A...")
		nodeB.Adapter.SendMessage("A", "Hello from Node B!")

	} else {
		fmt.Println("Invalid Node ID. Use A or B.")
		return
	}

	// Keep the program running
	wg.Wait()
}
