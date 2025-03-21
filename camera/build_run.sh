#!/bin/bash

# Build and run the binary locally

# Define variables
BINARY_PATH=./sudotest
REMOTE_USER=root
REMOTE_HOST=192.168.0.141
REMOTE_PATH=/usr/bin/
env GOOS=linux GOARCH=arm GOARM=7  go build -o sudotest main.go
# Copy the binary to the remote host
scp $BINARY_PATH $REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH || { echo "SCP failed!"; exit 1; }

