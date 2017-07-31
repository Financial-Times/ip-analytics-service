package queue

import (
	"fmt"
	"log"
	"time"

	"github.com/financial-times/ip-events-service/config"
	"github.com/satori/go.uuid"
	"github.com/streadway/amqp"
)

// Publisher for producing on queue
type Publisher struct {
	Queue   amqp.Queue
	Channel *amqp.Channel
}

// Publish publishes messages to queue
func (p *Publisher) Publish(body string, contentType string) error {
	err := p.Channel.Publish(
		"",           // exchange
		p.Queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: contentType,
			Body:        []byte(body),
			MessageId:   uuid.NewV4().String(),
			Timestamp:   time.Now(),
		})

	if err != nil {
		return err
	}
	log.Printf(" [x] Sent to %s", p.Queue.Name)
	return nil
}

// NewPublisher returns a new Publisher bound to a ch/queue
func NewPublisher(ch *amqp.Channel, cfg *config.Config) (*Publisher, error) {
	queueName := cfg.RabbitHost
	if queueName == "" {
		return nil, fmt.Errorf("RabbitHost is empty")
	}
	q, err := Declare(queueName, ch)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		Channel: ch,
		Queue:   q,
	}, nil
}
