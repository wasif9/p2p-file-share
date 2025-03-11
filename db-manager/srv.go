package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

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

// Handles http requests to the route '/records/{name}'
// Only defined for GET and DELETE
func createRecordHandler(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	recordHandler := func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		name := strings.TrimPrefix(r.URL.Path, "/api/"+VERSION+"/records/")

		switch r.Method {
		case http.MethodGet:
			record, err := querySingleRecord(db, name)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			err = json.NewEncoder(w).Encode(record)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case http.MethodDelete:
			err = deleteRecord(db, name)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		default:
			http.Error(w, fmt.Sprintf("Method %s not allowed.", r.Method), http.StatusMethodNotAllowed)
		}

	}

	return recordHandler
}

// Handles http requests to the route '/records'
// Only defined for POST
func createRecordsHandler(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	recordsHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodPost:
			var newManifest Manifest

			err = json.NewDecoder(r.Body).Decode(&newManifest)
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid request: %s", err.Error()), http.StatusBadRequest)
				return
			}

			createdManifest, err := insertRecord(db, &newManifest)
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
		default:
			http.Error(w, fmt.Sprintf("Method %s not allowed.", r.Method), http.StatusMethodNotAllowed)
		}
	}
	return recordsHandler
}

func killHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Kill request recieved. Shutting down server.")
	os.Exit(0)
}

// TODO: move this to common package so that client can use too
type Heartbeat struct {
	Index       int           `json:"node-index"`
	Uptime      time.Duration `json:"uptime"`
	Utilization int           `json:"utilization"`
}

func heartbeatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}

	err := json.NewEncoder(w).Encode(&Heartbeat{
		Index:       index,
		Uptime:      GetUptime(),
		Utilization: rand.Int() % 100,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
