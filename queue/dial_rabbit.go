package queue

import (
	"github.com/financial-times/ip-analytics-service/config"
	"github.com/streadway/amqp"
)

func DialRabbit(ap config.RabbitAddressProvider) (*amqp.Connection, chan *amqp.Error, error) {
	ad, err := ap.GetAddress()
	if err != nil {
		return nil, nil, err
	}
	conn, err := amqp.Dial(ad.String())
	if err != nil {
		return nil, nil, err
	}
	connErr := conn.NotifyClose(make(chan *amqp.Error))
	return conn, connErr, nil
}
