package queue

import (
	"github.com/streadway/amqp"
)

// Declare declares a queue on a channel
func Declare(queueName string, ch *amqp.Channel) (amqp.Queue, error) {
	q, err := ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	return q, err
}
