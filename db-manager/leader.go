package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"syscall"
	"time"

	types "github.com/wasif9/p2p-file-share/pkg/models"
)

var leaderAddr string = "http://localhost:99"
var status string = "none"

func monitorLeader() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	for true {
		time.Sleep(time.Second * 4)
		// every 4 seconds, make sure the leader is alive
		log.Printf("\tchecking in on node %d\n", leaderIndex)

		resp, err := http.Get(leaderAddr + "/api/v1/heartbeat")
		if errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ECONNABORTED) || errors.Is(err, syscall.ECONNRESET) {
			log.Println("\tleader DOWN!!!‼️‼️‼️‼️", err)

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

		log.Printf("leader says he is %d\n", heartbeat.Index)

	}

}

func election() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("we'll use the power of democracy")
	status = "candidate"
	client := &http.Client{Timeout: time.Second * 2}

	// Send Election message to other servers
	for _, peerNodeConfig := range allConfigs {
		if peerNodeConfig.Index <= cfg.Index {
			continue
		}

		// contact all nodes with higher index
		log.Println("sending a get...")
		resp, err := client.Get(fmt.Sprintf("http://%s:%s/api/v1/election/%d", peerNodeConfig.Host, peerNodeConfig.Port, cfg.Index))
		if err != nil {
			log.Printf("\tNo Response: %s\n", err.Error())
			log.Printf("\t%s must be dead\n", peerNodeConfig.Host)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("%s:%s is alive!\n", peerNodeConfig.Host, peerNodeConfig.Port)
		log.Printf("%s: %s\n", resp.Status, string(body))
		status = "loser"

	}

	if status == "candidate" {
		status = "leader"
		leaderAddr = net.JoinHostPort(cfg.Host, cfg.Port)
		log.Println("\tI AM THE WINNER!!! ♓♓")
		// leaderElected()
	}

	// continue to monitor the leader's heartbeat
	// go monitorLeader()
	log.Println("Done election().")
}

// func leaderElected() {

// 	// create message body
// 	jsonData, err := json.Marshal(leaderIndex)
// 	if err != nil {
// 		log.Println("Error encoding JSON:", err)
// 		return
// 	}

// 	// Send messages to other nodes
// 	for i := 0; i < len(nodeArr); i++ {
// 		if nodeArr[i].Index == node.Index {
// 			continue
// 		}
// 		postReq := fmt.Sprintf("http://%s:%s/api/v1/leader?%d", nodeArr[i].IP, nodeArr[i].Port, i)
// 		// postReq := "http://" + nodeArr[i].IP + ":" + nodeArr[i].Port + "/api/v1/leader"
// 		req, err := http.NewRequest("POST", postReq, bytes.NewBuffer(jsonData))
// 		if err != nil {
// 			log.Println("Error creating POST request:", err)
// 			continue
// 		}
// 		req.Header.Set("Content-Type", "application/json")

// 		resp, err := http.DefaultClient.Do(req)
// 		if err != nil {
// 			log.Println("Error sending POST request:", err)
// 			continue
// 		}
// 		defer resp.Body.Close()

// 		respSer, err := io.ReadAll(resp.Body)
// 		if err != nil {
// 			log.Println("Error reading response:", err)
// 			continue
// 		}

// 		log.Println("Resp Status:", resp.Status)
// 		log.Println("Resp Body:", string(respSer))
// 	}

// 	fmt.Println("sent leader elected message to all nodes")
// }
