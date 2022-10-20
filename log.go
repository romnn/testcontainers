package testcontainers

import (
	"context"
	"log"

	"github.com/testcontainers/testcontainers-go"
)

type LogCollector struct {
	LogChan   chan testcontainers.Log
	container testcontainers.Container
}

// Accept ...
func (logger *LogCollector) Accept(l testcontainers.Log) {
	logger.LogChan <- l
}

// Stop ...
func (logger *LogCollector) Stop() {
	logger.container.StopLogProducer()
	close(logger.LogChan)
}

// LogToStdout ...
func (logger *LogCollector) LogToStdout() {
	for {
		message, ok := <-logger.LogChan
		if !ok {
			return
		}
		log.Print(string(message.Content))
	}

}

// StartLogger ...
func StartLogger(ctx context.Context, c testcontainers.Container) (LogCollector, error) {
	logger := LogCollector{
		LogChan:   make(chan testcontainers.Log, 10),
		container: c,
	}

	err := c.StartLogProducer(ctx)
	if err == nil {
		c.FollowOutput(&logger)
	}
	return logger, err
}
