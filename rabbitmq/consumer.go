package rabbitmq

import (
	"github.com/streadway/amqp"
)

// ConsumerOptions ...
type ConsumerOptions struct {
	ExchangeName       string
	QueueName          string
	ExchangeRoutingKey string
	QueueLength        int64
}

// SetupConsumer ...
func SetupConsumer(options ConsumerOptions, ch *amqp.Channel) *amqp.Channel {
	// Define the arguments to configure a queue
	args := make(amqp.Table)
	// Set a maximum queue length
	args["x-max-length"] = options.QueueLength
	// Configure the Queue
	q, err := ch.QueueDeclare(
		options.QueueName, // name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		args,              // arguments
	)
	failOnError(err, "Failed to declare a queue")

	// Binding queue to specified exchange with routing key
	err = ch.QueueBind(
		q.Name,                     // queue name
		options.ExchangeRoutingKey, // routing key
		options.ExchangeName,       // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")

	return ch
}
