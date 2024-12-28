package main

import (
	"context"
	"log"
	"time"

	"github.com/IBM/sarama"
	tc "github.com/romnn/testcontainers"
	tckafka "github.com/romnn/testcontainers/kafka"
)

func main() {
	ctx := context.Background()
	container, err := tckafka.Start(ctx, tckafka.Options{
		// you could use latest here
		KafkaImageTag:     "5.4.10",
		ZookeeperImageTag: "3.8.0",
	})
	if err != nil {
		log.Fatalf("failed to start container: %v", err)
	}
	defer container.Terminate(ctx)

	// start logger
	if false {
		kafkaLogger, err := tc.StartLogger(ctx, container.Kafka.Container)
		if err != nil {
			log.Printf("failed to start logger for kafka: %v", err)
		} else {
			defer kafkaLogger.Stop()
			go kafkaLogger.LogToStdout()
		}
		zkLogger, err := tc.StartLogger(ctx, container.Zookeeper.Container)
		if err != nil {
			log.Printf("failed to start logger for zookeeper: %v", err)
		} else {
			defer zkLogger.Stop()
			go zkLogger.LogToStdout()
		}
	}

	topic := "my-topic"

	// prepare the consumer
	consumerOptions := tckafka.ConsumerOptions{
		Brokers: container.Kafka.Brokers,
		Group:   "TestConsumerGroup",
		Version: container.Kafka.Version,
		Topics:  []string{topic},
	}
	ctx, cancel := context.WithCancel(ctx)
	consumer, wg, err := consumerOptions.StartConsumer(ctx)
	if err != nil {
		log.Fatalf("failed to start the kafka consumer: %v", err)
	}

	// Prepare the producer
	producerOptions := tckafka.ProducerOptions{
		Brokers: container.Kafka.Brokers,
		Group:   "TestConsumerGroup",
		Version: container.Kafka.Version,
		Topics:  []string{topic},
	}
	producer, err := producerOptions.NewProducer()
	if err != nil {
		log.Fatalf("failed to produce events: %v", err)
	}

	messages := []string{
		"Message 1",
		"Message 2",
		"Message 3",
		"Message 4",
		"Message 5",
		"Message 6",
	}
	go func() {
		defer producer.Close()
		for _, msg := range messages {
			producer.Send(topic, msg, sarama.StringEncoder(msg))
		}
	}()

	defer consumer.Close()
	var received []string
	for len(received) < len(messages) {
		select {
		case msg := <-consumer.MessageChan:
			value := string(msg.Value)
			log.Printf("received message: %q", value)
			received = append(received, value)
		case <-time.After(10 * time.Second):
			log.Fatalf("sent %d messages, but received %d within timeout", len(messages), len(received))
		}
	}

	// signal consumer to exit its consume loop
	cancel()
	// wait for consumer to finish
	wg.Wait()

	log.Printf("received %v", received)
}
