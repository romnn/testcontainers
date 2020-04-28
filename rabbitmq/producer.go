package rabbitmq

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// Options for RabbitMQ...
type Options struct {
	Host               string
	Port               int64
	ExchangeName       string
	ExchangeRoutingKey string
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

// ConnectionURI ...
func (options Options) ConnectionURI() string {
	return fmt.Sprintf("amqp://guest:guest@%s:%d", options.Host, options.Port)
}

// Setup ...
func Setup(options Options) (*amqp.Connection, *amqp.Channel) {
	// 1. Create the connection to RabbitMQ
	rabbitmqURI := options.ConnectionURI()
	conn, err := amqp.Dial(rabbitmqURI)
	failOnError(err, "Failed to connect to RabbitMQ")

	// Initialize a channel for the connection
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	// Configure the exchange for the channel
	err = ch.ExchangeDeclare(
		options.ExchangeName, // name
		"direct",             // type
		true,                 // durable
		false,                // auto-deleted
		false,                // internal
		false,                // no-wait
		nil,                  // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	return conn, ch
}

// Publish ...
func Publish(message string, options Options, ch *amqp.Channel, count int) {
	body := message
	for i := 0; i < count; i++ {
		err := ch.Publish(
			options.ExchangeName,       // exchange
			options.ExchangeRoutingKey, // routing key
			false,                      // mandatory
			false,                      // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})
		failOnError(err, "Failed to publish a message")
	}
}
