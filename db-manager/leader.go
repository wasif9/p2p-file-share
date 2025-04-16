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
	"gorm.io/gorm"
)

var leaderIndex int = -1

func monitorLeader(db *gorm.DB) {
	for {
		time.Sleep(time.Second * 10)

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
		log.Printf("leader %d timestamp: %v\n", heartbeat.Index, heartbeat.Timestamp)

		timestamp := getTimestamp(db)

		if heartbeat.Timestamp > timestamp {
			log.Printf("leader heartbeat is newer than mine (%v > %v)\n", heartbeat.Timestamp, timestamp)
			catchup(timestamp)
		}

		if heartbeat.Timestamp < timestamp {
			log.Printf("my timestamp is newer than leader's (%v < %v), initiating election",
				heartbeat.Timestamp, timestamp)
			election()
			continue
		}
	}
}

func catchup(myTimestamp int) {
	log.Println("catching up to leader...")
	client := &http.Client{Timeout: time.Second * 2}

	resp, err := client.Post(
		fmt.Sprintf("http://%s/api/v1/catchup?timestamp=%d", allConfigs[leaderIndex].Address, myTimestamp),
		"", bytes.NewBuffer([]byte(strconv.Itoa(cfg.Index))),
	)
	if err != nil {
		log.Printf("failed to contact leader %d: %s", leaderIndex, err)
		return
	}
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("unexpected response from leader %d: %s: %s\n", leaderIndex, resp.Status, string(respBytes))
		return
	}
	var manifests []types.Manifest
	err = json.Unmarshal(respBytes, &manifests)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("got %d manifests from leader %d\n", len(manifests), leaderIndex)
	for _, manifest := range manifests {
		// post all to self
		manifestBytes, err := json.Marshal(manifest)
		if err != nil {
			log.Println(err)
			continue
		}
		payload := bytes.NewBuffer(manifestBytes)
		resp, err := client.Post(fmt.Sprintf("http://%s/api/v1/manifests", cfg.Address),
			"application/json",
			payload)
		if err != nil {
			log.Printf("error forwarding request to self, %s\n", err)
			continue
		}
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("error reading response from self: %s", err)
			continue
		}
		if resp.StatusCode != http.StatusCreated {
			log.Printf("error response from self, %s: %s\n", resp.Status, string(respBytes))
			continue
		}
		log.Printf("self acked\n")
	}
}

func reverseLeaderCatchup(db *gorm.DB, leaderAddress string) error {
	timestamp := getTimestamp(db)

	// Get leader's timestamp first
	client := &http.Client{Timeout: time.Second * 2}
	resp, err := client.Get(fmt.Sprintf("http://%s/api/v1/heartbeat", leaderAddress))
	if err != nil {
		return fmt.Errorf("failed to contact leader: %s", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %s", err)
	}

	var leaderHeartbeat types.Heartbeat
	err = json.Unmarshal(respBytes, &leaderHeartbeat)
	if err != nil {
		return fmt.Errorf("error parsing heartbeat: %s", err)
	}

	// If my timestamp is higher than leader's, initiate reverse catchup
	if timestamp > leaderHeartbeat.Timestamp {
		log.Printf("My timestamp (%d) is higher than leader's (%d), initiating reverse catchup",
			timestamp, leaderHeartbeat.Timestamp)

		// Query manifests that the leader is missing
		var manifests []types.Manifest
		err = db.Where("timestamp > ?", leaderHeartbeat.Timestamp).Find(&manifests).Error
		if err != nil {
			return fmt.Errorf("error querying newer manifests: %s", err)
		}

		log.Printf("Sending %d manifests to leader for reverse catchup", len(manifests))

		// Send these manifests to the leader via the reverse catchup endpoint
		manifestsBytes, err := json.Marshal(manifests)
		if err != nil {
			return fmt.Errorf("error serializing manifests: %s", err)
		}

		resp, err := client.Post(
			fmt.Sprintf("http://%s/api/v1/reverse-catchup", leaderAddress),
			"application/json",
			bytes.NewBuffer(manifestsBytes),
		)
		if err != nil {
			return fmt.Errorf("error sending manifests to leader: %s", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBytes, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("leader rejected reverse catchup: %s: %s",
				resp.Status, string(respBytes))
		}

		log.Printf("Reverse catchup to leader successful")
	}

	return nil
}

func election() {
	log.Println("begin election")

	client := &http.Client{Timeout: time.Second * 2}

	timestamps := make([]int, len(allConfigs))
	for i := range timestamps {
		timestamps[i] = -1
	}

	// poll everyone for their heartbeat to get the latest timestamp
	for _, peerNodeConfig := range allConfigs {
		resp, err := client.Get(fmt.Sprintf("http://%s/api/v1/heartbeat", peerNodeConfig.Address))
		if err != nil {
			log.Printf("failed to contact %d: %s", peerNodeConfig.Index, err)
			continue
		}
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			log.Printf("unexpected response from %d: %s: %s\n", peerNodeConfig.Index, resp.Status, string(respBytes))
			continue
		}
		var heartbeat types.Heartbeat
		err = json.Unmarshal(respBytes, &heartbeat)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("node %d timestamp: %v\n", heartbeat.Index, heartbeat.Timestamp)
		timestamps[heartbeat.Index] = heartbeat.Timestamp
	}

	// find the node with the highest timestamp
	highestTimestamp := int(0)
	highestIndex := cfg.Index
	for nodeIndex, timestamp := range timestamps {
		if timestamp > highestTimestamp { // the node with higher timestamp wins
			highestTimestamp = timestamp
			highestIndex = nodeIndex
		} else if timestamp == highestTimestamp {
			if nodeIndex < highestIndex { // tie break by the node index id when two nodes have the same timestamps
				highestTimestamp = timestamp
				highestIndex = nodeIndex
			}
		}
	}
	if highestIndex == -1 {
		log.Fatal("no nodes found")
	}
	log.Printf("node %d has the highest timestamp: %v\n", highestIndex, highestTimestamp)
	leaderIndex = highestIndex
	log.Printf("new leader is %d\n", leaderIndex)

	notifyFollowers(leaderIndex)
	notifyReverseProxy(leaderIndex)
}

func notifyFollowers(leaderIndex int) {
	log.Println("notifying followers...")
	client := &http.Client{Timeout: time.Second * 1}

	for _, peerNodeConfig := range allConfigs {
		resp, err := client.Post(
			fmt.Sprintf("http://%s/api/v1/leader",
				peerNodeConfig.Address),
			"", bytes.NewBuffer([]byte(strconv.Itoa(leaderIndex))),
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

func notifyReverseProxy(leaderIndex int) {
	resp, err := http.Post(fmt.Sprintf("http://%s/leader?address=%s",
		rpConfig.Address, allConfigs[leaderIndex].Address), "", nil)
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

func getTimestamp(db *gorm.DB) int {

	var timestamp int

	err = db.Model(&types.Manifest{}).Select("COALESCE(MAX(timestamp), 0)").Scan(&timestamp).Error
	if err != nil {
		log.Fatal(err)
	}
	return timestamp
}
