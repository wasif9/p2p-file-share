#!/bin/bash

# Usage: ./run_peers.sh <num_peers> [peer_app_path]
NUM_PEERS=$1
PEER_APP_PATH=${2:-.}

# Check for input
if [ -z "$NUM_PEERS" ]; then
    echo "Usage: ./run_peers.sh <num_peers> [peer_app_path]"
    exit 1
fi

echo "[+] Cleaning and setting up $NUM_PEERS peer directories..."

# Loop to setup directories
for ((i = 1; i <= NUM_PEERS; i++)); do
    DIR="node${i}-dir"
    
    # Delete dir if it exists
    if [ -d "$DIR" ]; then
        rm -rf "$DIR"
    fi

    # Recreate the dir
    mkdir "$DIR"

    # Populate p1-dir with sample .txt files
    if [ $i -eq 1 ]; then
        echo "This is file A from p1" > "$DIR/file_a.txt"
        echo "This is file B from p1" > "$DIR/file_b.txt"
        echo "Another test file from p1" > "$DIR/test_log.txt"
    fi
done

echo "[+] Launching $NUM_PEERS nodes..."

# Launch each peer node in a new terminal
for ((i = 1; i <= NUM_PEERS; i++)); do
    DIR="node${i}-dir"
    gnome-terminal -- bash -c "
        cd $PEER_APP_PATH;
        echo '[Node $i] Running from $DIR...';
        go run . --data-dir=$DIR;
        exec bash"
done
