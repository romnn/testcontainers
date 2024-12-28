package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ConsumerOptions ...
type ConsumerOptions struct {
	ExchangeName       string
	QueueName          string
	ExchangeRoutingKey string
	QueueLength        int64
}

// SetupQueue ...
func (options *ConsumerOptions) SetupQueue(ch *amqp.Channel) error {
	// define the arguments to configure a queue
	args := make(amqp.Table)
	args["x-max-length"] = options.QueueLength

	// configure the queue
	queue, err := ch.QueueDeclare(
		options.QueueName, // name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		args,              // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %v", err)
	}

	// bind queue to specified exchange with routing key
	err = ch.QueueBind(
		queue.Name,                 // queue name
		options.ExchangeRoutingKey, // routing key
		options.ExchangeName,       // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %v", err)
	}

	return nil
}
