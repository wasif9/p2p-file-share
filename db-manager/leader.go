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

var leaderIndex = 0

func monitorLeader() {

	for true {
		time.Sleep(time.Second * 4)
		// every 4 seconds, make sure the leader is alive
		log.Printf("\tchecking in on node %d\n", leaderIndex)
		fmt.Println(node.Port)

		// !HACK:
		// this assumes that node x runs on localhost 808x.
		// In reality, we'll need to let each node know the address (host:post) of each of it's peers, or maybe just it's successor
		leaderAddr := "http://localhost:808" + strconv.Itoa(leaderIndex)

		resp, err := http.Get(leaderAddr + "/api/v1/heartbeat")
		if err != nil {
			log.Println("\tleader DOWN!!!‼️‼️‼️‼️", err)
			node.Status = "candidate"
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
	// TODO: when servers startup they need to send their IP/Port

	// Send Election message to other servers
	fmt.Printf("Before For loop: %s\n", node.Status)
	for i := 0; i < len(nodeArr); i++ {
		fmt.Println(nodeArr[i].Index)
		resp, err := http.Get("http://" + nodeArr[i].IP + ":" + nodeArr[i].Port + "/api/v1/election/" + strconv.Itoa(nodeArr[i].Index))
		if err != nil {
			log.Println("\tNo Response", err)
			continue
		}
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
		nodeArr[i] = *node_peer
		//nodeArr[i].IP, nodeArr[i].Port, nodeArr[i].Index, nodeArr[i].Status = node_peer.IP, node_peer.Port, node_peer.Index, node_peer.Status
	}

	// Handle responses from peer nodes
	for i := 0; i < len(nodeArr); i++ {
		if nodeArr[i].Status == "candidate" && i != node.Index {
			node.Status = "waiting"
		}
	}
	fmt.Println(node.Status)
	// If node status is candidate, then node elected leader
	if node.Status == "candidate" {
		node.Status = "leader"
		leaderIndex = node.Index
		leaderElected()
	}
	// wait for leader election to finish
	for node.Status == "waiting" {
		time.Sleep(100 * time.Millisecond)
	}

	// ....
	// continue to monitor the leader's heartbeat
	go monitorLeader()
}

func leaderElected() {
	fmt.Println("send leader elected message to all nodes")
}
