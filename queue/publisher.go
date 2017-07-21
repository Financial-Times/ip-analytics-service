package queue

import (
	"fmt"
	"log"
	"time"

	"github.com/financial-times/ip-events-service/config"
	"github.com/satori/go.uuid"
	"github.com/streadway/amqp"
)

type Publisher struct {
}

func (p *Publisher) Publish(body string, contentType string, ch *amqp.Channel, cfg config.Config) error {
	queueName := cfg.RabbitHost
	if queueName == "" {
		return fmt.Errorf("RabbitHost is empty")
	}

	q, err := ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		return err
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: contentType,
			Body:        []byte(body),
			MessageId:   uuid.NewV4().String(),
			Timestamp:   time.Now(),
		})

	if err != nil {
		return err
	}
	log.Printf(" [x] Sent to %s", queueName)
	return nil
}
