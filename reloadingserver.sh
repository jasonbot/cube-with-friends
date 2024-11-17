#! /bin/bash

# This is for development -- does hot-reloading when things change
#
# Need to run:
#
#  go install github.com/air-verse/air@latest
#

air --build.cmd "go build -o ./server cmd/server.go" --build.bin "./server"
