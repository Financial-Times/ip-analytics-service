package events

import (
	"log"
	"time"

	"github.com/financial-times/ip-events-service/hooks"
	"github.com/financial-times/ip-events-service/queue"
	"github.com/streadway/amqp"
)

// App has a queue for publishing events
type App struct {
	queueHost string
	events    chan *hooks.FormattedEvent
	conn      *amqp.Connection
	connErr   chan *amqp.Error
	connRetry int
	ex        chan struct{}
}

// NewEventsApp creates a new Event App
func NewEventsApp(host string, ch <-chan *hooks.FormattedEvent) *App {
	return &App{
		queueHost: host,
		events:    ch,
		connRetry: 2,
		ex:        make(chan struct{}, 1),
	}
}

func (e *App) connect() error {
	conn, connErr, err := queue.DialRabbit(e.queueHost)
	if err != nil {
		return err
	}
	e.conn = conn
	e.connErr = connErr
	return nil
}

// Run the app
func (e *App) Run() error {
	log.Println("Starting Event App...")

	if err := e.connect(); err != nil {
		return err
	}
	defer e.conn.Close()
	_, err := e.conn.Channel()
	if err != nil {
		return err
	}

	// Start consmuming and publishing

	// connection control flow
	for {
		select {
		case err, ok := <-e.connErr:
			if err != nil {
				log.Printf("Queue connection lost: %s", err)
			}
			if !ok {
				log.Printf("Waiting %d seconds before reconnect attempt", e.connRetry)
				time.Sleep(time.Duration(e.connRetry) * time.Second)
			}
			if err := e.connect(); err != nil {
				log.Printf("Can't reconnect: %s", err)
				break
			}

		case <-e.ex:
			return nil
		}
	}
}

// Stop the app
func (e *App) Stop() {
	close(e.ex)
}
