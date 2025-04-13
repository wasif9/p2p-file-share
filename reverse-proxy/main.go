package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	types "github.com/wasif9/p2p-file-share/pkg/models"
)

var address string
var leaderAddr string
var superConfiguration types.SuperConfig
var robin = 0
var n int

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

	superConfiguration = types.SuperConfig{}
	err = json.Unmarshal(bytes, &superConfiguration)
	if err != nil {
		log.Fatalln("failed to unmarshal configuration file: ", err)
	}

	address = superConfiguration.RpConfig.Address
	leaderAddr = "unset"

	n = len(superConfiguration.DbManagerConfigs)
}

func nextRoundRobin() string {
	client := &http.Client{Timeout: time.Second * 1}

	for {
		try := superConfiguration.DbManagerConfigs[robin].Address
		robin = (robin + 1) % n

		log.Printf("trying %s\n", try)
		_, err := client.Get(fmt.Sprintf("http://%s/api/v1/heartbeat", try))
		if err == nil {
			log.Println("worked")
			return try
		}
	}
}

func main() {

	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {

		// choose the next target
		var target string
		if r.Method == http.MethodPost {
			target = leaderAddr
		} else {
			target = nextRoundRobin()
		}

		// Forward request to target server
		proxyReq, err := http.NewRequest(r.Method, "http://"+target+r.URL.Path+"?"+r.URL.RawQuery, r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error creating proxy request: %s", err), http.StatusInternalServerError)
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
			http.Error(w, fmt.Sprintf("Error forwarding request: %s", err), http.StatusBadGateway)
			return
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Fatal("Reverse proxy error when close the forwarding request", err)
			}
		}()

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
			if _, err := fmt.Fprint(w, leaderAddr); err != nil {
				log.Fatal("Reverse proxy error when GET /leader", err)
			}
		case http.MethodPost:
			newAddress := r.URL.Query().Get("address")
			if newAddress == "" {
				http.Error(w, "You need to specify the 'address' query parameter", http.StatusBadRequest)
				return
			}

			leaderAddr = newAddress
			log.Printf("- Forwarding /api/* requests to %s\n", leaderAddr)
		default:
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

	})

	log.Printf("Starting proxy server on %s\n", address)
	log.Printf("- Forwarding /api/* requests to %s\n", leaderAddr)
	log.Fatal(http.ListenAndServe(address, nil))
}
