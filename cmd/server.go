package main

import (
	"context"
	"cube-with-friends/mcgalaxyrunner"
	"log"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	runcontext, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := mcgalaxyrunner.RunGalaxyServer(runcontext, &wg)
	if err != nil {
		log.Fatalf("Error spinning up mc galaxy: %v", err)
	}

	time.Sleep(5 * time.Second)

	cancel()
	wg.Wait()
}
