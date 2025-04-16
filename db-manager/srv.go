package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	types "github.com/wasif9/p2p-file-share/pkg/models"
	"gorm.io/gorm"
)

var startTime time.Time

// special func name that runs once at beginning
func init() {
	startTime = time.Now()
}

func GetUptime() time.Duration {
	return time.Since(startTime)
}

// Handles http requests to the route '/manifests/{name}'
// Only defined for GET and DELETE
func manifestHandler(db *gorm.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		name := strings.TrimPrefix(r.URL.Path, "/api/"+cfg.Version+"/manifests/")

		switch r.Method {
		case http.MethodGet:
			manifest, err := querySingleManifest(db, name)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			err = json.NewEncoder(w).Encode(manifest)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case http.MethodDelete:
			err = deleteManifest(db, name)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		default:
			http.Error(w, fmt.Sprintf("Method %s not allowed.", r.Method), http.StatusMethodNotAllowed)
		}

	}
}

// Handles http requests to the route '/manifests'
// Only defined for POST
func manifestsHandler(db *gorm.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodPost:
			var newManifest types.Manifest

			err = json.NewDecoder(r.Body).Decode(&newManifest)
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid request: %s", err.Error()), http.StatusBadRequest)
				return
			}

			if leaderIndex == cfg.Index {
				log.Println("propagating write to followers...")
				copies := propagate(newManifest)
				followerCount := len(allConfigs) - 1 // -1 for self
				log.Printf("%d/%d followers acked, expecting %d/%d writes including self", copies, followerCount, copies+1, followerCount+1)
				if (copies + 1) <= (followerCount+1)/2.0 { // majority
					log.Println("failed to get majority ack")
					http.Error(w, "failed to get majority ack", http.StatusInternalServerError)
					return
				}
			}

			createdManifest, err := insertManifest(db, &newManifest)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
			err = json.NewEncoder(w).Encode(createdManifest)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		case http.MethodGet: // returns all manifests
			var manifests []types.Manifest

			if prefix := r.URL.Query().Get("prefix"); prefix != "" {
				manifests, err = queryManifestByPrefix(db, prefix)
			} else {
				manifests, err = queryManifest(db)
			}

			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid request: %s", err.Error()), http.StatusBadRequest)
				return
			}

			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(manifests)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		default:
			http.Error(w, fmt.Sprintf("Method %s not allowed.", r.Method), http.StatusMethodNotAllowed)
		}
	}
}

func propagate(newManifest types.Manifest) int {
	client := &http.Client{Timeout: time.Second * 1}
	successes := 0

	manifestBytes, err := json.Marshal(newManifest)
	if err != nil {
		log.Println(err)
		return 0
	}

	for i, peerConfig := range allConfigs {
		if peerConfig == cfg {
			continue // dont forward to self obv
		}
		payload := bytes.NewBuffer(manifestBytes)

		resp, err := client.Post(fmt.Sprintf("http://%s/api/v1/manifests", peerConfig.Address),
			"application/json",
			payload)
		if err != nil {
			log.Printf("error forwarding request to %d, %s\n", i, err)
			continue
		}
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("error reading response from %d: %s", i, err)
			continue
		}
		if resp.StatusCode != http.StatusCreated {
			log.Printf("error response from %d, %s: %s\n", i, resp.Status, string(respBytes))
			continue
		}
		log.Printf("%d acked\n", i)
		successes += 1
	}

	return successes
}

func killHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Kill request recieved. Shutting down server.")
	os.Exit(0)
}

func heartbeatHandler(db *gorm.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(&types.Heartbeat{
			Index:       cfg.Index,
			Uptime:      GetUptime(),
			LeaderIndex: leaderIndex,
			Timestamp:   getTimestamp(db),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// Handles http requests to the route '/election/{Index}'
// Only defined for GET
func electionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	_, err := fmt.Fprint(w, "candidate")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	go election()
}

// Handles http requests to the route '/leader'
func leaderHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&leaderIndex) // updates the leaderIndex
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %s", err.Error()), http.StatusBadRequest)
		return
	}

	log.Printf("my new leader is %d", leaderIndex)

	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprint(w, "Successfully updated leader!"); err != nil {
		log.Fatal("Error when POST leader", err)
	}

}

func catchupHandler(db *gorm.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
			return
		}
		requesterTimestampString := r.URL.Query().Get("timestamp")
		if requesterTimestampString == "" {
			http.Error(w, "Missing timestamp query parameter", http.StatusBadRequest)
			return
		}
		requesterTimestamp, err := strconv.Atoi(requesterTimestampString)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid timestamp: %s", err.Error()), http.StatusBadRequest)
			return
		}

		// get own timestamp
		timestamp := getTimestamp(db)

		// read the requester's index from the body
		var requesterIndex int
		err = json.NewDecoder(r.Body).Decode(&requesterIndex)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %s", err.Error()), http.StatusBadRequest)
			return
		}
		log.Printf("%d is at timestamp %d, catching up to %d\n", requesterIndex, requesterTimestamp, timestamp)

		// query the database for all manifests with a timestamp greater than the requester's
		var manifests []types.Manifest
		err = db.Where("timestamp > ?", requesterTimestamp).Find(&manifests).Error
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %s", err.Error()), http.StatusBadRequest)
			return
		}

		// send the manifests back to the requester
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(manifests)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("sent %d manifests to %d\n", len(manifests), requesterIndex)
	}
}
