# p2p-file-share
# Two-Node Communication Over HTTP & Ngrok

## Overview
This project sets up a **two-node communication system** using **HTTP RPC** and **Ngrok** for tunneling. Nodes communicate over HTTP and can send messages to each other, even if they are on different machines or behind NAT/firewalls.

---

## Installation & Setup
### **1️ Install Go**
Ensure you have Go installed. Check version:
```sh
go version
```
If not installed, download from [Go's official site](https://go.dev/dl/).

### **2️ Install Ngrok (For Internet-Based Testing)**


## Running the Project
### **1️ Start Ngrok for Each Node**
Open separate terminals and start Ngrok for **Node A & Node B**.
```sh
ngrok http 8081  # For Node A
ngrok http 8082  # For Node B
```
**Copy the generated URLs and replace them in `main.go` before running the nodes.**

---

### **2️ Start Each Node**
Open separate terminals for **Node A and Node B**.

#### **Start Node A**
```sh
go run main.go node.go real_network.go network_adapter.go -id=A
```

#### **Start Node B (After Node A is Running)**
```sh
go run main.go node.go real_network.go network_adapter.go -id=B
```

**Expected Output:**
```
[Node A] Listening on localhost:8081
[Node B] Listening on localhost:8082
Node A sending message to B...
Response from B: Message received
[Node B] Received: Hello from A!
```

---

### **3️ Testing Communication (Using curl)**
To manually test message delivery:
```sh
curl -X POST -d "Hello Node B!" https://randomsubdomainB.ngrok.io/rpc/message
```
Expected response:
```
Message received
```

---

## Notes
- **Ensure that each node’s Ngrok URL is correctly updated in `main.go`.**
- **Start Node A first, then Node B to avoid connection failures.**
- **If a node fails to send messages, ensure Ngrok is running and replace the URLs accordingly.**

---




