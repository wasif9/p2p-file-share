package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	types "github.com/wasif9/p2p-file-share/pkg/models"
)

var leaderIndex = 1

func isLeader() bool {
	return leaderIndex == cfg.Index
}

func monitorLeader() {
	for true {
		time.Sleep(time.Second * 4)
		// every 4 seconds, make sure the leader is alive
		log.Printf("\tchecking in on node %d\n", leaderIndex)

		leaderAddr := "http://localhost:808" + strconv.Itoa(leaderIndex)

		resp, err := http.Get(leaderAddr + "/api/v1/heartbeat")
		if err != nil {
			log.Println("\tleader DOWN!!!‼️‼️‼️‼️", err)
			election()
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

		log.Printf("leader says he is %d\n", heartbeat.Index)
	}
}

func election() {
	fmt.Println("we'll use the power of democracy")
	leaderIndex = cfg.Index
	go monitorLeader()
}
