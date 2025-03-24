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

func monitorLeader() {

	for true {
		time.Sleep(time.Second * 4)
		// every 4 seconds, make sure the leader is alive
		log.Printf("\tchecking in on node %d\n", leaderIndex)

		// !HACK:
		// this assumes that node x runs on localhost 808x.
		// In reality, we'll need to let each node know the address (host:post) of each of it's peers, or maybe just it's successor
		leaderAddr := "http//:localhost:8090"
		if leaderIndex < len(nodeArr) {
			leaderAddr = "http://" + nodeArr[leaderIndex].IP + ":" + nodeArr[leaderIndex].Port
		}

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
	node.Status = "candidate"
	// TODO: election process, update leaderIndex local variable
	// TODO: when servers startup they need to send their IP/Port

	// Send Election message to other servers
	for i := 0; i < len(nodeArr); i++ {

		if nodeArr[i].Index > node.Index {
			resp, err := http.Get("http://" + nodeArr[i].IP + ":" + nodeArr[i].Port + "/api/v1/election/" + strconv.Itoa(node.Index))
			if err != nil {
				log.Println("\tNo Response", err)
				nodeArr[i].Status = "No Response"
				continue
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}

			node_peer := new(types.Node)

			err = json.Unmarshal(body, &node_peer)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("node responded with %s\n", node_peer.Status)
			if node_peer.Status == "candidate" {
				node.Status = "waiting"

			}
			nodeArr[i].Status = node_peer.Status

		}

	}

	// Handle responses from peer nodes
	//for i := 0; i < len(nodeArr); i++ {

	// If there's another candidate present and it's not this node then wait
	//	if nodeArr[i].Status == "candidate" && i != node.Index {
	//		node.Status = "waiting"
	//	}
	//}

	// If node status is candidate, then node elected leader
	if node.Status == "candidate" {
		node.Status = "leader"
		leaderIndex = node.Index
		leaderElected()
	}
	// wait for leader election to finish
	//for node.Status == "waiting" {
	//	time.Sleep(100 * time.Millisecond)
	//}

	// continue to monitor the leader's heartbeat
	go monitorLeader()
}

func leaderElected() {

	// create message body
	jsonData, err := json.Marshal(leaderIndex)
	if err != nil {
		log.Println("Error encoding JSON:", err)
		return
	}

	// Send messages to other nodes
	for i := 0; i < len(nodeArr); i++ {
		if nodeArr[i].Index == node.Index {
			continue
		}
		postReq := "http://" + nodeArr[i].IP + ":" + nodeArr[i].Port + "/api/v1/leader"
		req, err := http.NewRequest("POST", postReq, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Println("Error creating POST request:", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("Error sending POST request:", err)
			continue
		}
		defer resp.Body.Close()

		respSer, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("Error reading response:", err)
			continue
		}

		log.Println("Resp Status:", resp.Status)
		log.Println("Resp Body:", string(respSer))
	}

	fmt.Println("sent leader elected message to all nodes")
}
