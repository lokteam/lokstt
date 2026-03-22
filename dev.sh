#!/bin/bash

export LOKSTT_CONFIG="/tmp/lokstt_dev.json"
export LOKSTT_SOCKET="/tmp/lokstt_dev.sock"

if [ "$1" == "ui" ]; then
    go run main.go --settings
else
    go run main.go &
    DAEMON_PID=$!
    
    sleep 1
    
    go run client/main.go SETTINGS
    
    wait $DAEMON_PID
fi
