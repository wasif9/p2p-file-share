package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
)

type Configuration struct {
	Address     string `json:"address"`
	Index       int    `json:"index"`
	Version     string `json:"version"`
	Pg_host     string `json:"pg-host"`
	Pg_user     string `json:"pg-user"`
	Pg_password string `json:"pg-password"`
	Pg_database string `json:"pg-database"`
	Pg_port     string `json:"pg-port"`
}

var cfg Configuration
var allConfigs []Configuration

func getConfig() Configuration {

	if len(os.Args) < 2 {
		log.Fatal("config file not provided")
	}

	configFilePath := os.Args[1]

	bytes, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Fatal(errors.Join(errors.New("Failed to read server configuration file"), err))
	}

	err = json.Unmarshal(bytes, &allConfigs)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 2 {
		i, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("2nd argument <node-index> must be int. Got: '%s'", os.Args[2])
		}
		return allConfigs[i]
	}

	for _, config := range allConfigs {
		if isLocalIP(config.Address) {
			_, err = http.Get(fmt.Sprintf("http://%s/api/v1/heartbeat", config.Address))
			if err != nil { // address is not yet taken. needed for running multiple nodes on single machine
				return config
			}
		}
	}
	log.Fatalf("Could not find local available address in config list")
	return Configuration{}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg = getConfig()

	log.Printf("Loaded configuration %+v\n", cfg)
}

func isLocalIP(ipStr string) bool {
	// Parse the input IP address
	inputIP := net.ParseIP(ipStr)
	if inputIP == nil {
		fmt.Println("Invalid IP address")
		return false
	}

	// Get all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Error getting network interfaces:", err)
		return false
	}

	// Iterate through network interfaces
	for _, iface := range interfaces {
		// Get addresses for this interface
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		// Check each address
		for _, addr := range addrs {
			// Convert to IP network
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Compare IPs
			if ip != nil && ip.Equal(inputIP) {
				return true
			}
		}
	}

	return false
}
