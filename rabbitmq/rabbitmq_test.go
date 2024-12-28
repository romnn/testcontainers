package rabbitmq

import (
	"context"
	"sort"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/romnn/deepequal"
	tc "github.com/romnn/testcontainers"
)

// TestRabbitmq...
func TestRabbitmq(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := Start(ctx, Options{
		ImageTag: "3.11.2",
	})
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}
	defer container.Terminate(ctx)

	// start logger
	logger, err := tc.StartLogger(ctx, container.Container)
	if err != nil {
		t.Errorf("failed to start logger: %v", err)
	} else {
		defer logger.Stop()
		go logger.LogToStdout()
	}

	exchangeName := "exchange_name"
	exchangeRoutingKey := "0"

	// setup the producer
	producerOptions := ProducerOptions{
		Host:               container.Host,
		Port:               container.Port,
		ExchangeName:       exchangeName,
		ExchangeRoutingKey: exchangeRoutingKey,
	}

	rmqConn, rmqChan, err := producerOptions.SetupConnection()
	if err != nil {
		t.Fatalf("failed to setup producer: %v", err)
	}
	defer rmqConn.Close()
	defer rmqChan.Close()

	// setup the consumer
	consumerOptions := ConsumerOptions{
		ExchangeName:       exchangeName,
		QueueName:          "queue_name",
		ExchangeRoutingKey: exchangeRoutingKey,
		QueueLength:        100,
	}

	err = consumerOptions.SetupQueue(rmqChan)
	if err != nil {
		t.Fatalf("failed to setup queue: %v", err)
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
		t.Fatalf("failed to register consumer: %v", err)
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
				t.Errorf(`failed to publish message "%s": %v`, msg, err)
			}
		}
	}()

	var received []string
	for len(received) != len(messages) {
		select {
		case msg := <-consumeChan:
			received = append(received, string(msg.Body))
		case <-time.After(10 * time.Second):
			t.Fatalf("sent %d messages, but received %d within timeout", len(messages), len(received))
		}
	}

	sort.Sort(sort.StringSlice(messages))
	sort.Sort(sort.StringSlice(received))
	if equal, err := deepequal.DeepEqual(messages, received); !equal {
		t.Errorf("expected %v, but received %v: %v", messages, received, err)
	}
}
