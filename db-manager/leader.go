package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	types "github.com/wasif9/p2p-file-share/pkg/models"
)

var leaderIndex int = -1

func monitorLeader() {
	for true {
		time.Sleep(time.Second * 5)

		if leaderIndex == -1 {
			log.Println("leader index is unset, calling election")

			election()
			continue
		}

		resp, err := http.Get(fmt.Sprintf("http://%s/api/v1/heartbeat", allConfigs[leaderIndex].Address))
		if err != nil {
			log.Println("‼️ leader down,", err)

			election()
			continue
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
	log.Println("begin election")

	winner := true // assume I am the biggest node index alive
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
			log.Printf("↳ No response from %d: %s\n", peerNodeConfig.Index, err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("%d is alive!\n", peerNodeConfig.Index)
		log.Printf("%s: %s\n", resp.Status, string(body))
		winner = false // there is a bigger node index alive. That node will continue the process and we will head from the winner eventually
		return
	}

	// no bigger node indeces were alive
	if winner {
		leaderIndex = cfg.Index
		log.Printf("✅ I, %d am the winner\n", cfg.Index)
		notifyFollowers()
		notifyReverseProxy()
	}
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
			log.Printf("notification message to %d failed: %s", peerNodeConfig.Index, err)
			continue
		}
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		if resp.StatusCode == http.StatusOK {
			log.Printf("node %d ack'ed notification", peerNodeConfig.Index)

		} else {
			log.Printf("node %d says %s: %s", peerNodeConfig.Index, resp.Status, string(respBytes))
		}
	}
}

func notifyReverseProxy() {
	resp, err := http.Post(fmt.Sprintf("http://%s/leader?address=%s", rpConfig.Address, cfg.Address), "", nil)
	if err != nil {
		log.Fatal("failed to contact reverse-proxy: ", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("unexpected response from reverse-proxy: %s: %s\n", resp.Status, string(respBytes))
		return
	}

	log.Println("reverse-proxy successfully updated")
}
