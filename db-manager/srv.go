package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Handles http requests to the route '/records/{name}'
// Only defined for GET and DELETE
func recordHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	name := strings.TrimPrefix(r.URL.Path, "/api/"+VERSION+"/records/")

	switch r.Method {
	case http.MethodGet:
		record, err := replaceWithGarethsSelectMethod(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		err = json.NewEncoder(w).Encode(record)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodDelete:
		err = replaceWithGarethsDeleteMethod()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	default:
		http.Error(w, fmt.Sprintf("Method %s not allowed.", r.Method), http.StatusMethodNotAllowed)
	}
}

// Handles http requests to the route '/records'
// Only defined for POST
func recordsHandler(w http.ResponseWriter, r *http.Request) {
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
