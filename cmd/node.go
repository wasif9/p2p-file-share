package main

import (
	"fmt"
	"io"
	"net/http"
)

// Node represents a peer
type Node struct {
	ID      string
	Address string
	Adapter NetworkAdapter
}

// HandleMessage processes incoming messages
func (n *Node) HandleMessage(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	fmt.Printf("[Node %s] Received: %s\n", n.ID, string(body))
	w.Write([]byte("Message received"))
}

// StartServer runs the node's HTTP server
func (n *Node) StartServer() {
	mux := http.NewServeMux() // Create a new ServeMux (isolated HTTP router)
	mux.HandleFunc("/rpc/message", n.HandleMessage)

	fmt.Println("[Node", n.ID, "] Listening on", n.Address)
	server := &http.Server{
		Addr:    n.Address,
		Handler: mux, // Use the custom router instead of the default
	}
	server.ListenAndServe()
}
