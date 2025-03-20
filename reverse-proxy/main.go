package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func main() {
	// Define addresses
	sourceAddr := "localhost:8080"
	targetAddr := "http://localhost:8081"

	// Create a handler function
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check if the path starts with /api
		if strings.HasPrefix(r.URL.Path, "/api") {
			// Forward request to target server
			proxyReq, err := http.NewRequest(r.Method, targetAddr+r.URL.Path+"?"+r.URL.RawQuery, r.Body)
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
		} else {
			// For non-API requests, return "Hello World"
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("Hello World"))
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	// Start the server
	fmt.Printf("Starting proxy server on %s\n", sourceAddr)
	fmt.Printf("- Forwarding /api/* requests to %s\n", targetAddr)
	fmt.Printf("- Returning 'Hello World' for all other requests\n")
	log.Fatal(http.ListenAndServe(sourceAddr, nil))
}
