#!/bin/bash

# Build and run the binary locally

# Define variables
BINARY_PATH=./sudocam
REMOTE_USER=blackbox
REMOTE_HOST=100.67.161.14
REMOTE_PATH=release/
env GOOS=linux GOARCH=arm64 go build -o sudocam
ssh $REMOTE_USER@$REMOTE_HOST "pkill -f sudocam" 

# Copy the binary to the remote host
scp $BINARY_PATH $REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH || { echo "SCP failed!"; exit 1; }
scp "./assets/post.json" $REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH || { echo "SCP failed!"; exit 1; }

# Execute the binary on the remote host
ssh $REMOTE_USER@$REMOTE_HOST "cd $REMOTE_PATH  && chmod +x sudocam && export PION_LOG_DEBUG=all && ./sudocam "
