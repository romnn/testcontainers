package main

import (
	"context"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	tc "github.com/romnn/testcontainers"
	tcrabbitmq "github.com/romnn/testcontainers/rabbitmq"
)

func main() {
	container, err := tcrabbitmq.Start(context.Background(), tcrabbitmq.Options{
		ImageTag: "3.11.2", // you could use latest here
	})
	if err != nil {
		log.Fatalf("failed to start container: %v", err)
	}
	defer container.Terminate(context.Background())

	// start logger
	logger, err := tc.StartLogger(context.Background(), container.Container)
	if err != nil {
		log.Printf("failed to start logger: %v", err)
	} else {
		defer logger.Stop()
		go logger.LogToStdout()
	}

	exchangeName := "exchange_name"
	exchangeRoutingKey := "0"

	// setup the producer
	producerOptions := tcrabbitmq.ProducerOptions{
		Host:               container.Host,
		Port:               container.Port,
		ExchangeName:       exchangeName,
		ExchangeRoutingKey: exchangeRoutingKey,
	}

	rmqConn, rmqChan, err := producerOptions.SetupConnection()
	if err != nil {
		log.Fatalf("failed to setup producer: %v", err)
	}
	defer rmqConn.Close()
	defer rmqChan.Close()

	// setup the consumer
	consumerOptions := tcrabbitmq.ConsumerOptions{
		ExchangeName:       exchangeName,
		QueueName:          "queue_name",
		ExchangeRoutingKey: exchangeRoutingKey,
		QueueLength:        100,
	}

	err = consumerOptions.SetupQueue(rmqChan)
	if err != nil {
		log.Fatalf("failed to setup queue: %v", err)
	}

	// consume queue
	consumeChan, err := rmqChan.Consume(
		consumerOptions.QueueName,          // queue
		consumerOptions.ExchangeRoutingKey, // consumer
		false,                              // auto ack
		false,                              // exclusive
		false,                              // no local
		false,                              // no wait
		nil,                                // args
	)
	if err != nil {
		log.Fatalf("failed to register consumer: %v", err)
	}

	// publish some messages
	messages := []string{"message1", "message2", "message3"}
	go func() {
		for _, msg := range messages {
			err := rmqChan.Publish(
				producerOptions.ExchangeName,       // exchange
				consumerOptions.ExchangeRoutingKey, // routing key
				false,                              // mandatory
				false,                              // immediate
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(msg),
				})
			if err != nil {
				log.Fatalf(`failed to publish message "%s": %v`, msg, err)
			}
		}
	}()

	var received []string
	for len(received) != len(messages) {
		select {
		case msg := <-consumeChan:
			received = append(received, string(msg.Body))
		case <-time.After(10 * time.Second):
			log.Fatalf("sent %d messages, but received %d within timeout", len(messages), len(received))
		}
	}
	log.Printf("received %v", received)
}
