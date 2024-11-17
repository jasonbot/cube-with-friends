# Cube: With Friends!

A minimal-effort server to run a locally hosted ClassiCube session.

This should work in Linux or MacOS.

## Building

```shell
# Get the embedded resources not hosted in this repo
./getresources.sh

# Build the exe
go build -o server cmd/server.go

# Or just run it where you are
go run cmd/main.go
```
