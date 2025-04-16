#!/bin/bash

# Usage: ./run_peers.sh <num_peers>
NUM_PEERS=$1

# Check for input
if [ -z "$NUM_PEERS" ]; then
    echo "Usage: ./run_peers.sh <num_peers>"
    exit 1
fi

echo "[+] Launching $NUM_PEERS nodes..."

# Launch each peer node in a new terminal
for ((i = 1; i <= NUM_PEERS; i++)); do
    gnome-terminal -- bash -c "
        echo '[Node $i] Running';
        go run ./... ../superconfig.json;
        exec bash"
done
