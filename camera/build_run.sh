#!/bin/bash

# Build and run the binary locally

# Define variables
BINARY_PATH=./blackbox
REMOTE_USER=blackbox
REMOTE_HOST=100.67.161.14
REMOTE_PATH=release/
env GOOS=linux GOARCH=arm64 go build -o blackbox
ssh $REMOTE_USER@$REMOTE_HOST "pkill -f blackbox" 

# Copy the binary to the remote host
scp $BINARY_PATH $REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH || { echo "SCP failed!"; exit 1; }

# Execute the binary on the remote host
ssh $REMOTE_USER@$REMOTE_HOST "cd $REMOTE_PATH  && chmod +x blackbox && export PION_LOG_DEBUG=all && ./blackbox "
