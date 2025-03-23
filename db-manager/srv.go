package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
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
	return time.Now().Sub(startTime)
}

// Handles http requests to the route '/manifests/{name}'
// Only defined for GET and DELETE
func createManifestHandler(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	manifestHandler := func(w http.ResponseWriter, r *http.Request) {

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

	return manifestHandler
}

// Handles http requests to the route '/manifests'
// Only defined for POST
func createManifestsHandler(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	manifestsHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodPost:
			var newManifest types.Manifest

			err = json.NewDecoder(r.Body).Decode(&newManifest)
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid request: %s", err.Error()), http.StatusBadRequest)
				return
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
	return manifestsHandler
}

func killHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Kill request recieved. Shutting down server.")
	os.Exit(0)
}

func heartbeatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(&types.Heartbeat{
		Index:       cfg.Index,
		Uptime:      GetUptime(),
		Utilization: rand.Int() % 100,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Handles http requests to the route '/election/{Index}'
// Only defined for POST
func electionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	index, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/"+cfg.Version+"/election/"))

	if index > node.Index {
		node.Status = "follower"
	} else {
		node.Status = "candidate"
	}

	err := json.NewEncoder(w).Encode(&types.Node{
		IP:     node.IP,
		Port:   node.Port,
		Index:  node.Index,
		Status: node.Status,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

// Handles http requests to the route '/leader'
func leaderHandler() func(w http.ResponseWriter, r *http.Request) {
	leaderHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
			return
		}

		var newLeaderIndex int

		err = json.NewDecoder(r.Body).Decode(&newLeaderIndex)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %s", err.Error()), http.StatusBadRequest)
			return
		}

		// udpate leader index and change node status
		leaderIndex = newLeaderIndex
		node.Status = "follower"

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Success!")

	}

	return leaderHandler
}
