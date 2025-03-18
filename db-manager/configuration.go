package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
)

type Configuration struct {
	Index       int    `json:"index"`
	Port        string `json:"port"`
	Version     string `json:"version"`
	Pg_host     string `json:"pg-host"`
	Pg_user     string `json:"pg-user"`
	Pg_password string `json:"pg-password"`
	Pg_database string `json:"pg-database"`
	Pg_port     string `json:"pg-port"`
}

var cfg Configuration

func init() {
	if len(os.Args) < 2 {
		log.Fatal("config file not provided")
	}

	configFilePath := os.Args[1]

	bytes, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Fatal(errors.Join(errors.New("Failed to read server configuration file"), err))
	}
	_ = bytes

	err = json.Unmarshal(bytes, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Loaded configuration %+v from %s\n", cfg, configFilePath)
}
