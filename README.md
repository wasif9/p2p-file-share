# p2p-file-share

## Overview
This project sets up a **two-node communication system** using **HTTP RPC**. Nodes communicate over HTTP and can send messages to each other.

---

## Installation & Setup
### **Install Go**
Ensure you have Go installed. Check version:
```sh
go version
```
If not installed, download from [Go's official site](https://go.dev/dl/).


## Running the Project

#### **Start Node A**
```sh
go run ./cmd -id=A
```

#### **In a New Terminal, Start Node B (ctrl+shift+5 in vscode)**
```sh
go run ./cmd -id=B
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

### **Testing Communication Using curl**
To manually test message delivery:
```sh
curl -X POST -d "My test message 🐱" http://localhost:8081/rpc/message
```
Expected response:
```
Message received
```
