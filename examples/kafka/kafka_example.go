package main

import (
	"context"
	"fmt"

	"github.com/Shopify/sarama"
	tc "github.com/romnnn/testcontainers"
	tckafka "github.com/romnnn/testcontainers/kafka"
	log "github.com/sirupsen/logrus"
)

func run() string {
	// Start kafka container
	kafkaC, Config, zkC, network, err := tckafka.StartKafkaContainer(context.Background(), tckafka.ContainerOptions{
		ContainerOptions: tc.ContainerOptions{
			// If you want to customize the container request
			/*
				ContainerRequest: testcontainers.ContainerRequest{
					Networks: []string{networkName},
				},
			*/
			// If you want to view the logs
			// CollectLogs: true,
		},
	})
	if err != nil {
		log.Fatalf("Failed to start the kafka container: %v", err)
	}
	defer network.Remove(context.Background())

	/*
		// If CollectLogs: true
		go func() {
			for {
				msg := <-Config.Log.MessageChan
				log.Info(msg)
			}
		}()
	*/

	// Prepare the consumer
	kcCtx, cancel := context.WithCancel(context.Background())
	kc, wg, err := tckafka.ConsumeGroup(kcCtx, tckafka.ConsumerOptions{
		Brokers: Config.Brokers,
		Group:   "TestConsumerGroup",
		Version: Config.KafkaVersion,
		Topics:  []string{"my-topic"},
	})
	if err != nil {
		log.Fatalf("Failed to start the kafka consumer: %v", err)
	}

	// Prepare the producer
	topic := "my-topic"
	kpCtx := context.Background()
	kp, err := tckafka.CreateProducer(kpCtx, tckafka.ProducerOptions{
		Brokers: Config.Brokers,
		Group:   "TestConsumerGroup",
		Version: Config.KafkaVersion,
		Topics:  []string{topic},
	})
	if err != nil {
		log.Fatalf("Cannot produce events: %v", err)
	}

	result := make(chan string)
	defer func() {
		if err := kp.Close(); err != nil {
			log.Errorf("Failed to close producer: %v", err)
		}
		if err := kc.Close(); err != nil {
			log.Errorf("Failed to close consumer: %v", err)
		}
		kafkaC.Terminate(context.Background())
		zkC.Terminate(context.Background())
	}()

	// Produce
	go func() {
		log.Info("Producer starting")
		for sent := 0; sent < 20; sent++ {
			kp.Send(topic, "my-message", sarama.StringEncoder(fmt.Sprintf("Message #%d", sent)))
		}
		kp.Send(topic, "end-message", sarama.StringEncoder("Consuming and producing works!"))
		log.Info("Producer done")
	}()

	// Consume
	go func() {
		log.Info("Consumer starting")
		var received int
		for msg := range kc.Messages {
			key := string(msg.Key)
			value := string(msg.Value)
			log.Infof("Received %s: %s", key, value)
			received++
			if key == "end-message" {
				// Never stop reading, only signal to stop
				result <- value
			}
		}
		log.Info("Consumer exited")
	}()

	final := <-result
	cancel()
	wg.Wait()
	return final
}

func main() {
	fmt.Println(run())
}
