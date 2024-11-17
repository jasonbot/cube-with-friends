package main

import (
	"context"
	"cube-with-friends/httpserver"
	"cube-with-friends/mcgalaxyrunner"
	"log"
	"os"
	"os/signal"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	runcontext, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			log.Println("Asking program to end gracefully...")
			cancel()
		}
	}()

	issuecommand, err := mcgalaxyrunner.RunGalaxyServer(cancel, runcontext, &wg)
	if err != nil {
		log.Fatalf("Error spinning up MCGalaxy: %v", err)
	}

	err = httpserver.ServeHttp(issuecommand, cancel, runcontext, &wg)
	if err != nil {
		log.Fatalf("Error spinning up MCGalaxy: %v", err)
	}

	cancel()
	wg.Wait()
}
