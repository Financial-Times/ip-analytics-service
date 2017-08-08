package queue

import (
	"context"
	"log"

	"github.com/streadway/amqp"
)

// Session composes amqp.Connection with amqp.Channel
type Session struct {
	*amqp.Connection
	*amqp.Channel
}

// Close tears down connection and takes channel with it
func (s Session) Close() error {
	if s.Connection == nil {
		return nil
	}
	return s.Connection.Close()
}

// Redial continually connects to host, exits when not possible
func Redial(ctx context.Context, url string, queueName string) chan chan Session {
	if queueName == "" {
		log.Fatalf("No queueName provided")
	}

	sessions := make(chan chan Session)

	go func() {
		sess := make(chan Session)
		defer close(sessions)

		for {
			select {
			case sessions <- sess:
			case <-ctx.Done():
				log.Println("shuttin down session factory")
				return
			}

			conn, err := amqp.Dial(url)
			if err != nil {
				log.Fatalf("cannot (re)dial: %v: %q", err, url)
			}

			ch, err := conn.Channel()
			if err != nil {
				log.Fatalf("cannot create channel: %v", err)
			}

			_, err = Declare(queueName, ch)
			if err != nil {
				log.Fatalf("cannot declare queue: %v", err)
			}

			select {
			case sess <- Session{conn, ch}:
			case <-ctx.Done():
				log.Println("shutting down new session")
				return
			}
		}
	}()

	return sessions
}
