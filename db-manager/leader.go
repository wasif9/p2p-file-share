package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"syscall"
	"time"

	types "github.com/wasif9/p2p-file-share/pkg/models"
)

var leaderIndex int = -1
var status string = "none"

func monitorLeader() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	for true {
		time.Sleep(time.Second * 10)
		// every 4 seconds, make sure the leader is alive
		log.Printf("checking in on leader %d\n", leaderIndex)

		resp, err := http.Get(fmt.Sprintf("http://%s/api/v1/heartbeat", allConfigs[leaderIndex].Address))
		if errors.Is(err, syscall.ECONNREFUSED) ||
			errors.Is(err, syscall.ECONNABORTED) ||
			errors.Is(err, syscall.ECONNRESET) ||
			err != nil { // makes aboce checks useless

			log.Println("‼️‼️‼️‼️ leader DOWN!!!‼️‼️‼️‼️", err)

			// election()
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		var heartbeat types.Heartbeat

		err = json.Unmarshal(body, &heartbeat)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("leader %d uptime: %v\n", heartbeat.Index, heartbeat.Uptime)
	}

}

func election() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("begin election")
	status = "candidate"
	client := &http.Client{Timeout: time.Second * 2}

	// Send Election message to other servers
	for i, peerNodeConfig := range allConfigs {
		if peerNodeConfig.Index <= cfg.Index {
			continue
		}

		// contact all nodes with higher index
		log.Printf("checking if %d is alive...\n", i)
		resp, err := client.Get(fmt.Sprintf("http://%s/api/v1/election/%d",
			peerNodeConfig.Address, cfg.Index))
		if err != nil {
			log.Printf("↳ No response from %d\n", peerNodeConfig.Index)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("%d is alive!\n", peerNodeConfig.Index)
		log.Printf("%s: %s\n", resp.Status, string(body))
		status = "loser"

	}

	// no other candidates were alive
	if status == "candidate" {
		status = "leader"
		leaderIndex = cfg.Index
		log.Printf("I, %d am the winner\n", cfg.Index)
		notifyFollowers()
	}

	// continue to monitor the leader's heartbeat
	log.Println("Done election.")
	go monitorLeader()
}

func notifyFollowers() {
	log.Println("notifying followers...")
	for _, peerNodeConfig := range allConfigs {
		resp, err := http.Post(
			fmt.Sprintf("http://%s/api/v1/leader",
				peerNodeConfig.Address),
			"", bytes.NewBuffer([]byte(strconv.Itoa(cfg.Index))),
		)
		if err != nil {
			log.Printf("notification message to %d failed", peerNodeConfig.Index)
			continue
		}
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("node %d says: %s: '%s'", peerNodeConfig.Index, resp.Status, string(respBytes))
	}
}
