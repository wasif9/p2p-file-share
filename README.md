# Peerify

## Overview
This report presents **Peerify**, a **hybrid peer-to-peer file sharing system** that addresses the limitations of traditional centralized architectures while maintaining practical usability. Centralized file-sharing systems suffer from single points of failure, limited scalability, and high maintenance costs, while fully decentralized approaches face challenges in implementation complexity and consistency management. 

Peerify introduces a three-layer architecture combining a decentralized peer network with a structured central registry. At the P2P layer, peers exchange files directly using a **Kademlia Distributed Hash Table (DHT)** for efficient file chunk location, enhanced by a dedicated bootstrap node. The orchestration layer features a reverse proxy that balances request distribution, while the central registry layer employs a primary-backup replication model with a consensus-based leader election protocol for metadata management. 

Our implementation demonstrates robust fault tolerance with automatic leader failover, eventual consistency through timestamp-based synchronization, and majority-write consensus for data durability. The system allows users to reliably upload and download files of any type while ensuring manifests are consistently replicated across distributed database nodes. 

Performance analysis confirms the system's resilience to node failures with k-1 fault tolerance in the P2P network and ⌊ (k - 1) / 2⌋ tolerance in the database layer. While maintaining these distributed characteristics, Peerify provides a seamless user experience through an intuitive interface that abstracts the underlying complexity of distributed systems operations.

---

## Installation & Setup
### **Install Go**
Ensure you have Go installed. Check version:
```sh
go version
```
If not installed, download from [Go's official site](https://go.dev/dl/).

---

### **Install GioUI for GUI**
```bash
go install gioui.org/cmd/gogio@latest
``` 
to install the latest [Gio UI](https://gioui.org/)

---

### **Install PostgreSQL**
Ensure you have PostgreSQL installed. Check version:
```sh
psql --version
```
If not installed, download from [PostgreSQL's official site](https://www.postgresql.org/).

### Creating the registry[x] database

```bash
psql -U postgres
```

```sql
CREATE DATABASE registry0;
```

---

### **Modify `superconfig.json`**
Change the ``address`` of each node so that those address can be reached

---

## Running the Project

#### **Install Dependencies**
```sh
go mod tidy
```

#### **Start Reverse Proxy**
```sh
cd ./reverse-proxy; go run ./... ../superconfig.json
```

#### **Start Database Manager**
```sh
cd ./db-manager; go run ./... ../superconfig.json 0
```
where ``0`` is the index of the database manager

**Note:** Each database manager connects to each own database. For example, manager ``0`` connect to database ``registry0``

---
#### **(Optional) Start Bootstrap Node**
We currently run our bootstrap node on Google Cloud Platform. Peers need to connect to a bootstrap node to join the P2P network. You can run your own bootstrap node.
```sh
cd ./bootstrap; go run ./...
```
Then modify the ``.env`` file in ``peer-app`` folder. Change the content to
```
BOOTSTRAP_ADDR=/ip4/xxx.xxx.xxx.xxx/tcp/4001/p2p/...
```
where ``BOOTSTRAP_ADDR`` is the multiaddrs shown in your bootstrap node

---
#### **Start GUI**
```sh
cd ./peer-app; go run ./... ../superconfig.json
```

---

Now peers can upload and download files 🌟 
