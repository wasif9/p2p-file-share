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

func monitorLeader() {
	for true {
		time.Sleep(time.Second * 4)
		// every 4 seconds, make sure the leader is alive
		log.Printf("\tchecking in on node %d\n", leaderIndex)

		// !HACK:
		// this assumes that node x runs on localhost 808x.
		// In reality, we'll need to let each node know the address (host:post) of each of it's peers, or maybe just it's successor
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
	// TODO: election process, update leaderIndex local variable
	// ....
	// continue to monitor the leader's heartbeat
	go monitorLeader()
}
