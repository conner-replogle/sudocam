#!/bin/bash

# Build and run the binary locally

# Define variables
BINARY_PATH=./sudotest
REMOTE_USER=root
REMOTE_HOST=192.168.0.141
REMOTE_PATH=/usr/bin/

PASSWORD=luckfox
export CGO_CFLAGS="-Ideps/include"
export CGO_LDFLAGS="-Ldeps/lib -lrknnmrt"
env GOOS=linux GOARCH=arm GOARM=7  go build -o sudotest main.go

# Calculate the local binary hash
LOCAL_HASH=$(sha256sum $BINARY_PATH | awk '{print $1}')

# Get the remote binary hash
REMOTE_HASH=$(sshpass -p "$PASSWORD" ssh $REMOTE_USER@$REMOTE_HOST "sha256sum $REMOTE_PATH/sudotest 2>/dev/null | awk '{print \$1}'")
sshpass -p "$PASSWORD" ssh $REMOTE_USER@$REMOTE_HOST "killall -9 sudotest"

# Check if the hashes are the same
if [ "$LOCAL_HASH" == "$REMOTE_HASH" ]; then
    echo "Hashes are the same, not uploading the binary"
else
    echo "Hashes are different, uploading the binary"
    # Kill the process if it's running
    sshpass -p "$PASSWORD" ssh $REMOTE_USER@$REMOTE_HOST "killall -9 sudotest"
    # Copy the binary to the remote host
    sshpass -p "$PASSWORD" scp $BINARY_PATH $REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH || { echo "SCP failed!"; exit 1; }
fi

# Run the binary on the remote host
sshpass -p "$PASSWORD" ssh $REMOTE_USER@$REMOTE_HOST "$REMOTE_PATH/sudotest --config dev-config.json"