package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"
)

const (
	maxRetries       = 20
	retryIntervalSec = 5
)

// ConsumerOptions ...
type ConsumerOptions struct {
	Brokers []string
	Group   string
	Version string
	Topics  []string
}

// Consumer ...
type Consumer struct {
	ready    chan error
	Messages chan *sarama.ConsumerMessage
	client   sarama.ConsumerGroup
}

// Cleanup ...
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim ...
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		log.Debugf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
		consumer.Messages <- message
		session.MarkMessage(message, "")
	}
	return nil
}

// Setup ...
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	consumer.ready <- nil
	return nil
}

// Close ...
func (consumer *Consumer) Close() error {
	return consumer.client.Close()
}

func newConsumerGroup(options ConsumerOptions, config *sarama.Config) (sarama.ConsumerGroup, error) {
	var attempt int
	for {
		client, err := sarama.NewConsumerGroup(options.Brokers, options.Group, config)
		if err != nil {
			if attempt >= maxRetries {
				return nil, fmt.Errorf("Failed to start kafka consumer: %s", err.Error())
			}
			attempt++
			log.Infof("Failed to connect: %s. (Attempt %d of %d)", err.Error(), attempt, maxRetries)
			time.Sleep(time.Duration(retryIntervalSec) * time.Second)
			continue
		}
		return client, nil
	}
}

// ConsumeGroup ...
func ConsumeGroup(ctx context.Context, options ConsumerOptions) (*Consumer, *sync.WaitGroup, error) {
	config := sarama.NewConfig()
	// config.Consumer.Return.Errors = true

	c := &Consumer{
		ready:    make(chan error, 2),
		Messages: make(chan *sarama.ConsumerMessage),
	}

	version, err := sarama.ParseKafkaVersion(options.Version)
	if err != nil {
		return c, nil, fmt.Errorf("Error parsing Kafka version: %v", err)
	}
	config.Version = version

	c.client, err = newConsumerGroup(options, config)
	if err != nil {
		return c, nil, fmt.Errorf("Error creating consumer group client: %v", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// Consume should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := c.client.Consume(ctx, options.Topics, c); err != nil {
				c.ready <- err
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				log.Debug("Context canceled")
				c.ready <- err
				break
			}
			c.ready = make(chan error, 2)
		}
		log.Debug("Consumer loop exiting")
	}()

	// Wait until the consumer has been set up
	if err = <-c.ready; err != nil {
		return c, wg, fmt.Errorf("Error setting up consumer: %v", err)
	}

	log.Debug("Sarama consumer up and running!...")
	return c, wg, nil
}
