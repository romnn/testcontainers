package kafka

import (
	// "context"
	"fmt"
	"log"
	"time"

	"github.com/Shopify/sarama"
	// "github.com/cenkalti/backoff/v4"
	// log "github.com/sirupsen/logrus"
)

// ProducerOptions ...
type ProducerOptions struct {
	Brokers []string
	Group   string
	Version string
	Topics  []string
}

// Producer ...
type Producer struct {
	DataCollector     sarama.SyncProducer
	AccessLogProducer sarama.AsyncProducer
}

// Close ...
func (p *Producer) Close() error {
	if err := p.DataCollector.Close(); err != nil {
		return fmt.Errorf("failed to shut down data collector: %v", err)
	}
	if err := p.AccessLogProducer.Close(); err != nil {
		return fmt.Errorf("failed to shut down access log producer: %v", err)
	}
	return nil
}

// Send ...
func (p *Producer) Send(topic string, pkey string, entry sarama.Encoder) {
	p.AccessLogProducer.Input() <- &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(pkey),
		Value: entry,
	}
}

// func (p *Producer) forBroker(brokerList []string) (Producer, error) {
// func newProducer(brokerList []string) (Producer, error) {
// 	collector, err := newDataCollector(brokerList)
// 	if err != nil {
// 		return nil, err
// 	}
// 	producer, err := newAccessLogProducer(brokerList)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return Producer{
// 		DataCollector:     collector,
// 		AccessLogProducer: producer,
// 	}, nil
// }

func newDataCollector(brokerList []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	// wait for all in-sync replicas to ack the message
	config.Producer.RequiredAcks = sarama.WaitForAll
	// retry up to 10 times to produce the message
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true
	return sarama.NewSyncProducer(brokerList, config)
}

func newAccessLogProducer(brokerList []string) (sarama.AsyncProducer, error) {
	config := sarama.NewConfig()
	// only wait for the leader to ack
	config.Producer.RequiredAcks = sarama.WaitForLocal
	// compress messages
	config.Producer.Compression = sarama.CompressionSnappy
	// flush batches every 500ms
	config.Producer.Flush.Frequency = 500 * time.Millisecond
	producer, err := sarama.NewAsyncProducer(brokerList, config)
	if err != nil {
		return nil, err
	}
	go func() {
		for err := range producer.Errors() {
			log.Printf("failed to write access log entry: %v\n", err)
		}
	}()
	return producer, err
}

// StartProducer ...
func (options *ProducerOptions) NewProducer() (*Producer, error) {
	// func (options *ProducerOptions) NewProducer(ctx context.Context) (*Producer, error) {
	// var producer Producer
	collector, err := newDataCollector(options.Brokers)
	if err != nil {
		return nil, err
	}
	producer, err := newAccessLogProducer(options.Brokers)
	if err != nil {
		return nil, err
	}
	return &Producer{
		DataCollector:     collector,
		AccessLogProducer: producer,
	}, nil

	// producer, err := newProducer(options.Brokers)
	// return producer, err
	// newProducer := func() error {
	// 	var err error
	// for {
	// 	select {
	// 	case <-ctx.Done():
	// 		return nil, fmt.Errorf("Failed to start kafka producer after %d attempts", attempt)
	// 	default:

	// 	// group, err = sarama.NewConsumerGroup(options.Brokers, options.Group, config)
	// 	return err
	// }

	// bo := backoff.NewExponentialBackOff()
	// bo.Multiplier = 1.1
	// bo.InitialInterval = 2 * time.Second
	// bo.MaxElapsedTime = 2 * time.Minute
	// err := backoff.Retry(newProducer, bo)
	// return producer, err

	// var attempt int
	// for {
	// 	select {
	// 	case <-ctx.Done():
	// 		return nil, fmt.Errorf("Failed to start kafka producer after %d attempts", attempt)
	// 	default:
	// 		producer, err := Producer{}.forBroker(options.Brokers)
	// 		if err != nil {
	// 			if attempt >= maxRetries {
	// 				return nil, fmt.Errorf("Failed to start kafka producer: %v", err)
	// 			}
	// 			attempt++
	// 			log.Infof("Failed to connect: %v. (Attempt %d of %d)", err, attempt, maxRetries)
	// 			time.Sleep(time.Duration(retryIntervalSec) * time.Second)
	// 		} else {
	// 			return producer, nil
	// 		}
	// 	}
	// }
}
