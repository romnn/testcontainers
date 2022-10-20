package kafka

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/cenkalti/backoff/v4"
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
	readyChan   chan error
	MessageChan chan *sarama.ConsumerMessage
	client      sarama.ConsumerGroup
}

// Cleanup ...
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim ...
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// Do not move the code below to a goroutine.
	// `ConsumeClaim` itself is called within a goroutine, see:
	for {
		select {
		case message := <-claim.Messages():
			consumer.MessageChan <- message
			session.MarkMessage(message, "")

			// Should return when `session.Context()` is done.
		case <-session.Context().Done():
			return nil
		}
	}
}

// Setup ...
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	consumer.readyChan <- nil
	return nil
}

// Close ...
func (consumer *Consumer) Close() error {
	return consumer.client.Close()
}

func (options *ConsumerOptions) newConsumerGroup(config *sarama.Config) (sarama.ConsumerGroup, error) {

	var group sarama.ConsumerGroup
	newConsumerGroup := func() error {
		var err error
		group, err = sarama.NewConsumerGroup(options.Brokers, options.Group, config)
		return err
	}

	bo := backoff.NewExponentialBackOff()
	bo.Multiplier = 1.1
	bo.InitialInterval = 2 * time.Second
	bo.MaxElapsedTime = 2 * time.Minute
	err := backoff.Retry(newConsumerGroup, bo)
	return group, err
}

// StartConsumer ...
func (options *ConsumerOptions) StartConsumer(ctx context.Context) (Consumer, *sync.WaitGroup, error) {
	config := sarama.NewConfig()

	consumer := Consumer{
		readyChan:   make(chan error), // , 2),
		MessageChan: make(chan *sarama.ConsumerMessage),
	}

	version, err := sarama.ParseKafkaVersion(options.Version)
	if err != nil {
		return consumer, nil, fmt.Errorf(`failed to parse kafka version "%q": %v`, options.Version, err)
	}
	config.Version = version

	consumer.client, err = options.newConsumerGroup(config)
	if err != nil {
		return consumer, nil, fmt.Errorf("error creating consumer group: %v", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// Consume should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := consumer.client.Consume(ctx, options.Topics, &consumer); err != nil {
				consumer.readyChan <- err
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
				// log.Println("context canceled")
				// consumer.readyChan <- err
				// break
			}
			// consumer.readyChan = make(chan error, 2)
			consumer.readyChan = make(chan error)
		}
		log.Println("consumer loop exiting")
	}()

	// Wait until the consumer has been set up
	<-consumer.readyChan

	// if err = <-consumer.readyChan; err != nil {
	// 	return consumer, wg, fmt.Errorf("error setting up consumer: %v", err)
	// }

	log.Println("consumer started")
	return consumer, wg, nil
}
