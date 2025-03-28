package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
)

type Configuration struct {
	Host        string `json:"host"`
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
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if len(os.Args) < 2 {
		log.Fatal("config file not provided")
	}

	configFilePath := os.Args[1]

	bytes, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Fatal(errors.Join(errors.New("Failed to read server configuration file"), err))
	}

	var cfgs []Configuration

	err = json.Unmarshal(bytes, &cfgs)
	if err != nil {
		log.Fatal(err)
	}

	for _, config := range cfgs {
		if isLocalIP(config.Host) {
			cfg = config
		}
	}

	log.Printf("Loaded configurations %+v from %s\n", cfgs, configFilePath)
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
