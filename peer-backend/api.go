package main

import (
	"log"
	"net/http"
)

func get(filename string) {
	resp, err := http.Get("http://localhost:8080/api/v1/records/" + filename)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Get responded with status %s\n", resp.Status)
	}

	
}
