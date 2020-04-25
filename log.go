package testcontainers

import (
	"context"

	"github.com/prometheus/common/log"
	"github.com/testcontainers/testcontainers-go"
)

// LogCollector ...
type LogCollector struct {
	MessageChan chan string
}

// Accept ...
func (c *LogCollector) Accept(l testcontainers.Log) {
	c.MessageChan <- string(l.Content)
}

func enableLogger(container testcontainers.Container, logger *LogCollector) {
	*logger = LogCollector{
		MessageChan: make(chan string),
	}

	if err := container.StartLogProducer(context.Background()); err != nil {
		log.Errorf("Failed to start log producer: %v", err)
		return
	}

	container.FollowOutput(logger)
	// User must call StopLogProducer() himself
}
