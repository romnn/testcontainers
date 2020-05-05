package main

import (
	"context"

	tcrabbitmq "github.com/romnnn/testcontainers/rabbitmq"
	log "github.com/sirupsen/logrus"
)

func run() int {
	// Start rabbitmq container
	rabbitmqCont, rabbitmqConf, err := tcrabbitmq.StartRabbitmqContainer(tcrabbitmq.ContainerOptions{})
	if err != nil {
		log.Fatalf("Failed to start rabbitMQ container: %v", err)
	}
	defer rabbitmqCont.Terminate(context.Background())

	exchangeName := "exchange_name"
	queueName := "queue_name"
	exchangeRoutingKey := "0"
	var queueLength int64 = 100

	options := tcrabbitmq.Options{
		Host:               rabbitmqConf.Host,
		Port:               rabbitmqConf.Port,
		ExchangeName:       exchangeName,
		ExchangeRoutingKey: exchangeRoutingKey,
	}

	// Setup RabbitMQ
	rmqConn, rmqCh := tcrabbitmq.Setup(options)
	defer rmqConn.Close()
	defer rmqCh.Close()

	consumerOptions := tcrabbitmq.ConsumerOptions{
		ExchangeName:       exchangeName,
		QueueName:          queueName,
		ExchangeRoutingKey: exchangeRoutingKey,
		QueueLength:        queueLength,
	}

	// Setup the RabbitMQConsumer
	consumerCh := tcrabbitmq.SetupConsumer(consumerOptions, rmqCh)
	defer consumerCh.Close()

	// Publish mock data
	publishCount := 40
	shortMessage := "short Message"
	longMessage := "This is a longer message"
	tcrabbitmq.Publish(shortMessage, options, rmqCh, publishCount/2)
	tcrabbitmq.Publish(longMessage, options, rmqCh, publishCount/2)

	// Consume queue
	msgs, err := consumerCh.Consume(
		consumerOptions.QueueName,          // queue
		consumerOptions.ExchangeRoutingKey, // consumer
		false,                              // auto ack
		false,                              // exclusive
		false,                              // no local
		false,                              // no wait
		nil,                                // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	done := make(chan bool)
	consumedCount := 0
	shortMessageCounter := 0

	go func() {
		for d := range msgs {
			log.Infof("%s \n", d.Body)
			if len(d.Body) < 20 {
				shortMessageCounter++
			}
			consumedCount++
			if consumedCount == publishCount {
				done <- true
				return
			}
		}
	}()
	<-done
	return shortMessageCounter
}

func main() {
	log.Infof("Found %d short messages in the queue", run())
}
