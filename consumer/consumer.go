package consumer

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/financial-times/ip-events-service/hooks"
	"github.com/financial-times/ip-events-service/queue"
)

// Client for sending events to different apps
type Client interface {
	Send(fe hooks.FormattedEvent) error
	IsInternal() bool
}

// Consume from message chan and sends to destination via Client
func Consume(msgs chan queue.Message, cs ...Client) error {

	for m := range msgs {
		var wg sync.WaitGroup
		doneChan := make(chan bool, 1)
		errChan := make(chan error, 1)

		fe := make([]hooks.FormattedEvent, 0)
		if err := json.Unmarshal(m.Body, &fe); err != nil {
			log.Println("Couldn't unmarshal body to send")
			log.Println(err)
			m.Response <- true
			continue
		}

		for _, e := range fe {
			wg.Add(1)

			go func(ev hooks.FormattedEvent, m queue.Message) {
				defer wg.Done()
				var err error
				for _, c := range cs {
					if !c.IsInternal() || ev.Internal { // check if only whitelisted for internal sending
						err = c.Send(ev)
					}
					if err != nil {
						errChan <- err
						break
					}
				}
			}(e, m)
		}

		go func() {
			wg.Wait()
			close(doneChan)
		}()

		select {
		case <-doneChan:
			log.Println("Sent")
			m.Response <- true
		case err := <-errChan:
			if err != nil {
				log.Println("Couldn't send")
				log.Println(err)
				m.Response <- false
			}
		}
	}

	return nil
}
