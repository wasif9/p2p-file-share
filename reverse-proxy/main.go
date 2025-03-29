package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var address string

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if len(os.Args) < 2 {
		log.Fatalln("usage: go run ./... <superconfiguration-filepath>")
	}
	configFilePath := os.Args[1]

	bytes, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Fatalln("failed to read configuration file: ", err)
	}
	_ = bytes
}

func main() {
	// Define addresses
	sourceAddr := "0.0.0.0:8080"
	leaderAddr := "http://localhost:8081" // TODO: !hardcoded

	// Create a handler function
	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		// Check if the path starts with /api
		// Forward request to target server
		proxyReq, err := http.NewRequest(r.Method, leaderAddr+r.URL.Path+"?"+r.URL.RawQuery, r.Body)
		if err != nil {
			http.Error(w, "Error creating proxy request", http.StatusInternalServerError)
			return
		}

		// Copy headers
		for name, values := range r.Header {
			for _, value := range values {
				proxyReq.Header.Add(name, value)
			}
		}

		// Send the request
		client := &http.Client{}
		resp, err := client.Do(proxyReq)
		if err != nil {
			http.Error(w, "Error forwarding request", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for name, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}

		// Copy status code and body
		w.WriteHeader(resp.StatusCode)
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			log.Fatal(err)
		}
	})

	http.HandleFunc("/leader", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			fmt.Fprint(w, leaderAddr)
		case http.MethodPost:
			newAddress := r.URL.Query().Get("address")
			if newAddress == "" {
				http.Error(w, "You need to specify the 'address' query parameter", http.StatusBadRequest)
				return
			}

			leaderAddr = newAddress
		default:
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

	})
	// Start the server
	log.Printf("Starting proxy server on %s\n", sourceAddr)
	log.Printf("- Forwarding /api/* requests to %s\n", leaderAddr)
	log.Fatal(http.ListenAndServe(sourceAddr, nil))
}
