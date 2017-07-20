package queue

import (
	"github.com/streadway/amqp"
)

func DialRabbit(ap config.RabbitAddressProvider) (*amqp.Connection, chan *amqp.Error, error) {
	add, err := ap.GetAddress()
	if err != nil {
		return nil, nil, err
	}
}
