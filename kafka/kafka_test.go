package kafka

import (
	"context"
	"testing"
	"time"
	"log"

	"github.com/Shopify/sarama"
	"github.com/romnn/deepequal"
	tc "github.com/romnn/testcontainers"
)

// TestKafka...
func TestKafka(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := Start(ctx, Options{
		KafkaImageTag:     "5.4.10",
		ZookeeperImageTag: "3.8.0",
	})
	if err != nil {
		t.Fatalf("failed to start kafka container: %v", err)
	}
	defer container.Terminate(ctx)

	// start logger
	kafkaLogger, err := tc.StartLogger(ctx, container.Kafka.Container)
	if err != nil {
		t.Errorf("failed to start logger for kafka: %v", err)
	} else {
		defer kafkaLogger.Stop()
		go kafkaLogger.LogToStdout()
	}
	zkLogger, err := tc.StartLogger(ctx, container.Zookeeper.Container)
	if err != nil {
		t.Errorf("failed to start logger for zookeeper: %v", err)
	} else {
		defer zkLogger.Stop()
		go zkLogger.LogToStdout()
	}

	topic := "my-topic"

	// prepare the consumer
	consumerOptions := ConsumerOptions{
		Brokers: container.Kafka.Brokers,
		Group:   "TestConsumerGroup",
		Version: container.Kafka.Version,
		Topics:  []string{topic},
	}
	ctx, cancel := context.WithCancel(ctx)
	consumer, wg, err := consumerOptions.StartConsumer(ctx)
	if err != nil {
		t.Fatalf("failed to start the kafka consumer: %v", err)
	}

	// Prepare the producer
	producerOptions := ProducerOptions{
		Brokers: container.Kafka.Brokers,
		Group:   "TestConsumerGroup",
		Version: container.Kafka.Version,
		Topics:  []string{topic},
	}
	// producer, err := CreateProducer(ctx, producerOptions)
	producer, err := producerOptions.NewProducer()
	if err != nil {
		t.Fatalf("failed to produce events: %v", err)
	}

	// done := make(chan string)
	// done := make(chan bool)
	// defer func() {
	// 	if err := kp.Close(); err != nil {
	// 		t.Errorf("Failed to close producer: %v", err)
	// 	}
	// 	if err := kc.Close(); err != nil {
	// 		t.Errorf("Failed to close consumer: %v", err)
	// 	}
	// 	container.Terminate(ctx)
	// 	// zkC.Terminate(ctx)
	// }()

	// Produce
	messages := []string{
		"Message 1",
		"Message 2",
		"Message 3",
		"Message 4",
		"Message 5",
		"Message 6",
	}
	// produced := 20
	go func() {
		defer producer.Close()
		// defer consumer.Close()
		for _, msg := range messages {
			producer.Send(topic, msg, sarama.StringEncoder(msg))
		}
		// for sent := 0; sent < produced-1; sent++ {
		// 	producer.Send(topic, "my-message", sarama.StringEncoder(fmt.Sprintf("Message #%d", sent)))
		// }
		// producer.Send(topic, "end-message", sarama.StringEncoder("Consuming and producing works!"))
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
			t.Fatalf("sent %d messages, but received %d within timeout", len(messages), len(received))
		}
		// for msg := range consumer.MessageChan {
		// msg := <-consumer.MessageChan
		// received = append(received, string(msg.Value))
	}
	// key := string(msg.Key)
	// value := string(msg.Value)
	// received++
	// if key == "end-message" {
	// 	// Never stop reading, only signal to stop
	// 	done <- value
	// }
	// }

	// Consume
	// var received int
	// go func() {
	// for range consumer.MessageChan {
	// 		// key := string(msg.Key)
	// 		// value := string(msg.Value)
	// 		received++
	// 		// if key == "end-message" {
	// 		// 	// Never stop reading, only signal to stop
	// 		// 	done <- value
	// 		// }
	// 	}
	// 	done <- true
	// }()

	// signal consumer to exit its consume loop
	cancel()
	// wait for consumer to finish
	wg.Wait()

	// kafka maintains ordering of messages, so no need to sort
	if equal, err := deepequal.DeepEqual(messages, received); !equal {
		t.Errorf("expected %v, but received %v: %v", messages, received, err)
	}
}
