#!/usr/bin/env bash


# Get the pid of the process that runs on port 8080

port=$(lsof -i :8080 | grep LISTEN | awk '{print $2}')
if [ -z "$port" ]; then
    echo "No process is running on port 8080"
else
    echo "Killing process $port"
    kill -3 "$port"
fi