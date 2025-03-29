package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"strconv"
)

type DbMgrConfig struct {
	Address     string `json:"address"`
	Index       int    `json:"index"`
	Version     string `json:"version"`
	Pg_host     string `json:"pg-host"`
	Pg_user     string `json:"pg-user"`
	Pg_password string `json:"pg-password"`
	Pg_database string `json:"pg-database"`
	Pg_port     string `json:"pg-port"`
}

type revProxyConfig struct {
	Address string `json:"address"`
}
type superConfig struct {
	revProxyConfig   `json:"reverse-proxy"`
	DbManagersConfig []DbMgrConfig `json:"db-managers"`
}

var cfg DbMgrConfig
var allConfigs []DbMgrConfig
var rpConfig revProxyConfig

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if len(os.Args) < 3 {
		log.Fatal("usage: go run ./... <config-file> <node-index>")
	}

	configFilePath := os.Args[1]

	bytes, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Fatal(errors.Join(errors.New("Failed to read server configuration file"), err))
	}

	mySuperConfig := superConfig{}
	err = json.Unmarshal(bytes, &mySuperConfig)
	if err != nil {
		log.Fatal(err)
	}

	allConfigs = mySuperConfig.DbManagersConfig
	rpConfig = mySuperConfig.revProxyConfig

	i, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatalf("2nd argument <node-index> must be int. Got: '%s'", os.Args[2])
	}
	cfg = allConfigs[i]

	log.Printf("Loaded configuration %+v\n", cfg)
}
