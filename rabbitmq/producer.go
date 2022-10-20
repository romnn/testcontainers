package rabbitmq

import (
	"fmt"

	"github.com/streadway/amqp"
)

// ProducerOptions for rabbitmq
type ProducerOptions struct {
	Host               string
	Port               int64
	ExchangeName       string
	ExchangeRoutingKey string
}

// ConnectionURI ...
func (options *ProducerOptions) ConnectionURI() string {
	return fmt.Sprintf("amqp://guest:guest@%s:%d", options.Host, options.Port)
}

// SetupConnnection ...
func (options *ProducerOptions) SetupConnection() (*amqp.Connection, *amqp.Channel, error) {
	// connect
	uri := options.ConnectionURI()
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to rmq: %v", err)
	}

	// initialize a channel for the connection
	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open connection channel: %v", err)
	}

	// configure the exchange for the channel
	err = ch.ExchangeDeclare(
		options.ExchangeName, // name
		"direct",             // type
		true,                 // durable
		false,                // auto-deleted
		false,                // internal
		false,                // no-wait
		nil,                  // arguments
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to declare exchange: %v", err)
	}

	return conn, ch, nil
}
