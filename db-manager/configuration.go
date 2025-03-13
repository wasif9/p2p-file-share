package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
)

type Configuration struct {
	index       int
	PORT        string
	VERSION     string
	PG_PASSWORD string
}

func Configure() Configuration {

	bytes, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatal(errors.Join(errors.New("Failed to read server configuration file"), err))
	}
	_ = bytes

	var cfg Configuration

	err = json.Unmarshal(bytes, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}
