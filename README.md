# Cube: With Friends!

A minimal-effort server to run a locally hosted ClassiCube session with the web interface enabled for a group of people on a local network.

This should work in Linux or MacOS. Requires you have mono installed via `{apt,dnf,brew} install mono`.

The final binary artifact (minus the Mono runtime) can be distributed and run anywhere, it has all its other dependencies baked in.

## Building

```shell
# Get the embedded resources not hosted in this repo
./getresources.sh

# Build the exe
go build -o server cmd/server.go

# Or just run it where you are
go run cmd/main.go
```
