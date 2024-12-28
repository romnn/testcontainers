package kafka

import (
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
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

// NewProducer ...
func (options *ProducerOptions) NewProducer() (*Producer, error) {
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
}
