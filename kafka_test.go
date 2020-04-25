package testcontainers

import (
	"context"
	"fmt"

	"testing"

	"github.com/Shopify/sarama"
	"github.com/romnnn/testcontainers/kafka"
)

// TestKafkaContainer ...
func TestKafkaContainer(t *testing.T) {
	t.Parallel()
	// Start kafka container
	kafkaC, kafkaConfig, zkC, network, err := StartKafkaContainer(KafkaContainerOptions{
		ContainerOptions: ContainerOptions{},
	})
	if err != nil {
		t.Fatalf("Failed to start the kafka container: %v", err)
	}
	defer network.Remove(context.Background())

	// Prepare the consumer
	kcCtx, cancel := context.WithCancel(context.Background())
	kc, wg, err := kafka.ConsumeGroup(kcCtx, kafka.ConsumerOptions{
		Brokers: kafkaConfig.Brokers,
		Group:   "TestConsumerGroup",
		Version: kafkaConfig.KafkaVersion,
		Topics:  []string{"my-topic"},
	})
	if err != nil {
		t.Fatalf("Failed to start the kafka consumer: %v", err)
	}

	// Prepare the producer
	topic := "my-topic"
	kpCtx := context.Background()
	kp, err := kafka.CreateProducer(kpCtx, kafka.ProducerOptions{
		Brokers: kafkaConfig.Brokers,
		Group:   "TestConsumerGroup",
		Version: kafkaConfig.KafkaVersion,
		Topics:  []string{topic},
	})
	if err != nil {
		t.Fatalf("Cannot produce events: %v", err)
	}

	done := make(chan string)
	defer func() {
		if err := kp.Close(); err != nil {
			t.Errorf("Failed to close producer: %v", err)
		}
		if err := kc.Close(); err != nil {
			t.Errorf("Failed to close consumer: %v", err)
		}
		kafkaC.Terminate(context.Background())
		zkC.Terminate(context.Background())
	}()

	// Produce
	produced := 20
	go func() {
		for sent := 0; sent < produced-1; sent++ {
			kp.Send(topic, "my-message", sarama.StringEncoder(fmt.Sprintf("Message #%d", sent)))
		}
		kp.Send(topic, "end-message", sarama.StringEncoder("Consuming and producing works!"))
	}()

	// Consume
	var received int
	go func() {
		for msg := range kc.Messages {
			key := string(msg.Key)
			value := string(msg.Value)
			received++
			if key == "end-message" {
				// Never stop reading, only signal to stop
				done <- value
			}
		}
	}()

	<-done
	cancel()
	wg.Wait()
	if received != produced {
		t.Fatalf("Produced %d messages but received %d messages", produced, received)
	}
}
