package testcontainers

import (
	"context"
	"testing"

	"github.com/romnnn/testcontainers/rabbitmq"
)


// TestRabbitmqContainer ...
func TestRabbitmqContainer(t *testing.T) {
	t.Parallel()
	// Start rabbitmq container
	rabbitmqCont, rabbitmqConf, err := StartRabbitmqContainer(RabbitmqContainerOptions{})
	if err != nil {
		t.Fatalf("Failed to start rabbitMQ container: %v", err)
	}
	defer rabbitmqCont.Terminate(context.Background())


    exchangeName := "exchange_name"
    queueName := "queue_name"
    exchangeRoutingKey := "0"
    var queueLength int64 = 100

	options := rabbitmq.RabbitMQOptions{
        Host: rabbitmqConf.Host,
        Port: rabbitmqConf.Port,
        ExchangeName: exchangeName,
        ExchangeRoutingKey: exchangeRoutingKey,
    }

	// Setup RabbitMQ
    rmqConn, rmqCh := rabbitmq.Setup(options)
    defer rmqConn.Close()
    defer rmqCh.Close()

    consumerOptions := rabbitmq.ConsumerOptions{
        ExchangeName: exchangeName,
        QueueName: queueName,
        ExchangeRoutingKey: exchangeRoutingKey,
        QueueLength: queueLength,
    }

	// Setup the RabbitMQConsumer
    consumerCh := rabbitmq.SetupConsumer(consumerOptions, rmqCh)
    defer consumerCh.Close()

	// Publish mock data
	publishCount := 15
	message := ""
	rabbitmq.Publish(message, options, rmqCh,  publishCount)

	// Consume queue
	 msgs, err := consumerCh.Consume(
          consumerOptions.QueueName, // queue
          consumerOptions.ExchangeRoutingKey,     // consumer
          false,   // auto ack
          false,  // exclusive
          false,  // no local
          false,  // no wait
          nil,    // args
     )
     if err != nil {
		t.Fatalf("Failed to register a consumer: %v", err)
	}

     forever := make(chan bool)
     consumedCount := 0

     go func() {
          for range msgs {
              consumedCount++
              if consumedCount > 10 {
                forever <- true
              }
          }
     }()
     <-forever

     if consumedCount != 11 {
        t.Fatalf("Expected 11 items in the database but got %d", consumedCount)
     }
}