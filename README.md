# Cube: With Friends!

A minimal-effort server to run a locally hosted ClassiCube session with the web interface enabled for a group of people on a local network.

This should work in Linux or MacOS. Requires you have mono installed via `{apt,dnf,brew} install mono` (or `mono-runtime`).

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

## Running

```shell

# Run directly from compilation step
$ go run cmd/main.go

# or after doingh `go build`
$ ./server
```

This will place all the necessary files in the `game/` folder, and run the server from that CWD. You can embed resources to 'overlay'; e.g. there is a custom `server.properties` we drop in place to disable authentication.

When the server runs, it will give you a list of (possible) URLs to share with friends to get this to work in the logs.
