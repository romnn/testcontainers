package rabbitmq

import (
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	amqp "github.com/rabbitmq/amqp091-go"
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

// SetupConnection ...
func (options *ProducerOptions) SetupConnection() (*amqp.Connection, *amqp.Channel, error) {
	var conn *amqp.Connection
	var ch *amqp.Channel
	uri := options.ConnectionURI()

	setupConnection := func() error {
		var err error

		conn, err = amqp.Dial(uri)
		if err != nil {
			return fmt.Errorf("failed to connect to rmq: %v", err)
		}

		// initialize a channel for the connection
		ch, err = conn.Channel()
		if err != nil {
			return fmt.Errorf("failed to open connection channel: %v", err)
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
			return fmt.Errorf("failed to declare exchange: %v", err)
		}
		return nil
	}

	bo := backoff.NewExponentialBackOff()
	bo.Multiplier = 1.2
	bo.InitialInterval = 5 * time.Second
	bo.MaxElapsedTime = 2 * time.Minute
	err := backoff.Retry(setupConnection, bo)
	return conn, ch, err
}
