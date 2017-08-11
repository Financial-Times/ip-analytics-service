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
			msgs <- Message{Body: msg.Body}
			// TODO send response chan and wait for confirmation before ack'in or nack'in
			sub.Ack(msg.DeliveryTag, false)
		}
	}
}

// Write used to write consumed message
func Write(w io.Writer) chan Message {
	msgs := make(chan Message)
	go func() {
		for msg := range msgs {
			fmt.Fprintln(w, string(msg.Body))
		}
	}()
	return msgs
}
