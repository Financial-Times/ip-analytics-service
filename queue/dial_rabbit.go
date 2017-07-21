package queue

import (
	"github.com/streadway/amqp"
)

func DialRabbit(rHost string) (*amqp.Connection, chan *amqp.Error, error) {
	conn, err := amqp.Dial(rHost)
	if err != nil {
		return nil, nil, err
	}
	connErr := conn.NotifyClose(make(chan *amqp.Error))
	return conn, connErr, nil
}
