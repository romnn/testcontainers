package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"
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

func newDataCollector(brokerList []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack the message
	config.Producer.Retry.Max = 10                   // Retry up to 10 times to produce the message
	config.Producer.Return.Successes = true
	return sarama.NewSyncProducer(brokerList, config)
}

func newAccessLogProducer(brokerList []string) (sarama.AsyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal       // Only wait for the leader to ack
	config.Producer.Compression = sarama.CompressionSnappy   // Compress messages
	config.Producer.Flush.Frequency = 500 * time.Millisecond // Flush batches every 500ms
	producer, err := sarama.NewAsyncProducer(brokerList, config)
	if err != nil {
		return nil, err
	}
	go func() {
		for err := range producer.Errors() {
			log.Warnf("Failed to write access log entry: %v", err)
		}
	}()
	return producer, err
}

// Close ...
func (p *Producer) Close() error {
	if err := p.DataCollector.Close(); err != nil {
		return fmt.Errorf("Failed to shut down data collector cleanly: %v", err)
	}
	if err := p.AccessLogProducer.Close(); err != nil {
		return fmt.Errorf("Failed to shut down access log producer cleanly: %v", err)
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

func (p Producer) forBroker(brokerList []string) (*Producer, error) {
	collector, err := newDataCollector(brokerList)
	if err != nil {
		return nil, err
	}
	producer, err := newAccessLogProducer(brokerList)
	if err != nil {
		return nil, err
	}
	return &Producer{
		DataCollector:     collector,
		AccessLogProducer: producer,
	}, nil
}

// CreateProducer ...
func CreateProducer(ctx context.Context, options ProducerOptions) (*Producer, error) {
	var attempt int
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("Failed to start kafka producer after %d attempts", attempt)
		default:
			producer, err := Producer{}.forBroker(options.Brokers)
			if err != nil {
				if attempt >= maxRetries {
					return nil, fmt.Errorf("Failed to start kafka producer: %v", err)
				}
				attempt++
				log.Infof("Failed to connect: %v. (Attempt %d of %d)", err, attempt, maxRetries)
				time.Sleep(time.Duration(retryIntervalSec) * time.Second)
			} else {
				return producer, nil
			}
		}
	}
}
