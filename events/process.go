package events

import (
	"github.com/financial-times/ip-events-service/config"
	"github.com/financial-times/ip-events-service/queue"
	"github.com/streadway/amqp"
)

// Publisher interface for rabbitmq publisher
type Publisher interface {
	Publish(body string, contentType string, ch *amqp.Channel, cfg *config.Config) error
}

// New returns a new Handler instance
func New(e <-chan *hooks.FormattedEvent, ch *amqp.Channel, cfg config.Config) (*Handler, error) {
	p, err := queue.NewPublisher(ch, cfg)
	if err != nil {
		return nil, err
	}
	h := &Handler{
		publisher: p,
		events:    e,
	}
	return h, h.start()
}

// Handler for consuming and publishing events
type Handler struct {
	publisher
	events chan *hooks.FormattedEvent
}

func (h *Handler) start() error {
	// run process as goroutine
	// listen on h.events
	// publish events to queue
}
