package main

import (
	"context"
	"log"
	"time"

	tc "github.com/romnn/testcontainers"
	tcredis "github.com/romnn/testcontainers/redis"
)

func run() {
	container, err := tcredis.Start(context.Background(), tcredis.Options{
		ImageTag: "7.0.5", // you could use latest here
	})
	if err != nil {
		log.Fatalf("failed to start container: %v", err)
	}
	defer container.Terminate(context.Background())

	// start logger
	logger, err := tc.StartLogger(context.Background(), container.Container)
	if err != nil {
		log.Fatalf("failed to start logger: %v", err)
	}
	defer logger.Stop()
	go func() {
		for {
			message, ok := <-logger.LogChan
			if !ok {
				return
			}
			log.Printf("custom log: %s", string(message.Content))
		}
	}()

	// collect logs for 4 seconds
	time.Sleep(4 * time.Second)
}

func main() {
	run()
}
