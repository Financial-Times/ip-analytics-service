package queue

import (
	"fmt"
	"io"
	"log"
)

// Consume takes messages from queue and publishes to kinesis
func Consume(sessions chan chan Session, msgs chan<- Message, queueName string) {
	for session := range sessions {
		sub := <-session

		deliveries, err := sub.Consume(queueName, "", false, false, false, false, nil)
		if err != nil {
			log.Printf("cannot consume from: %q, %v", queueName, err)
			return
		}

		log.Printf("Consuming...")

		for msg := range deliveries {
			confirm := make(chan bool, 1)
			msgs <- Message{Body: msg.Body, Response: confirm}
			ok := <-confirm
			if ok {
				sub.Ack(msg.DeliveryTag, false)
			} else {
				sub.Nack(msg.DeliveryTag, false, true)
			}
		}
	}
}

// Write used to write consumed message
func Write(w io.Writer) chan Message {
	msgs := make(chan Message)
	go func() {
		for m := range msgs {
			fmt.Fprintln(w, string(m.Body))
			m.Response <- true
		}
	}()
	return msgs
}
