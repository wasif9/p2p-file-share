package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"strconv"

	types "github.com/wasif9/p2p-file-share/pkg/models"
)

var cfg types.DbMgrConfig
var allConfigs []types.DbMgrConfig
var rpConfig types.RevProxyConfig

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

	mySuperConfig := types.SuperConfig{}
	err = json.Unmarshal(bytes, &mySuperConfig)
	if err != nil {
		log.Fatal(err)
	}

	allConfigs = mySuperConfig.DbManagerConfigs
	rpConfig = mySuperConfig.RpConfig

	i, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatalf("2nd argument <node-index> must be int. Got: '%s'", os.Args[2])
	}
	cfg = allConfigs[i]

	log.Printf("Loaded configuration %+v\n", cfg)
}
