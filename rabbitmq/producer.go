package rabbitmq

import (
        "fmt"

        "github.com/streadway/amqp"
	    log "github.com/sirupsen/logrus"
)

type RabbitMQOptions struct {
	Host                string
	Port                int64
	ExchangeName        string
	ExchangeRoutingKey  string
}

func failOnError(err error, msg string) {
        if err != nil {
                log.Fatalf("%s: %s", msg, err)
        }
}

func Setup(options RabbitMQOptions) (*amqp.Connection, *amqp.Channel){
  // 1. Create the connection to RabbitMQ
  url := fmt.Sprintf("amqp://guest:guest@%s:%d", options.Host, options.Port)
  conn, err := amqp.Dial(url)
  failOnError(err, "Failed to connect to RabbitMQ")

  // Initialize a channel for the connection
  ch, err := conn.Channel()
  failOnError(err, "Failed to open a channel")

  // Configure the exchange for the channel
  err = ch.ExchangeDeclare(
          options.ExchangeName, // name
          "direct",      // type
          true,          // durable
          false,         // auto-deleted
          false,         // internal
          false,         // no-wait
          nil,           // arguments
  )
  failOnError(err, "Failed to declare an exchange")

  return conn, ch
}

func Publish(message string, options RabbitMQOptions, ch *amqp.Channel, count int) {
    body := message
    for i := 0; i < count; i ++ {
        err := ch.Publish(
                options.ExchangeName,         // exchange
                options.ExchangeRoutingKey, // routing key
                false, // mandatory
                false, // immediate
                amqp.Publishing{
                        ContentType: "text/plain",
                        Body:        []byte(body),
                })
        failOnError(err, "Failed to publish a message")
    }
}
