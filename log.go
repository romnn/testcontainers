package testcontainers

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/testcontainers/testcontainers-go"
)

// LogCollector ...
type LogCollector struct {
	MessageChan chan string
	Mux         sync.Mutex
}

// Accept ...
func (c *LogCollector) Accept(l testcontainers.Log) {
	c.MessageChan <- string(l.Content)
}

// EnableLogger ...
func EnableLogger(container testcontainers.Container, logger *LogCollector) {
	/**logger = LogCollector{
		MessageChan: make(chan string),
		// mux: sync.Mutex{},
	}
	*/

	// logger.mux.Lock()
	// defer logger.mux.Unlock()

	if err := container.StartLogProducer(context.Background()); err != nil {
		logrus.Errorf("Failed to start log producer: %v", err)
		return
	}
	container.FollowOutput(logger)
	// User must call StopLogProducer() himself
}
