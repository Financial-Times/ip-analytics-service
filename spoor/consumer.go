package spoor

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/financial-times/ip-events-service/hooks"
	"github.com/financial-times/ip-events-service/queue"
)

// Consume from message chan and sends to spoor via Client
func Consume(msgs chan queue.Message, c *Client) error {

	for m := range msgs {
		var wg sync.WaitGroup
		doneChan := make(chan bool, 1)
		errChan := make(chan error, 1)

		fe := make([]hooks.FormattedEvent, 0)
		if err := json.Unmarshal(m.Body, &fe); err != nil {
			log.Printf("Couldn't unmarshal body to send to Spoor: %v", err)
			m.Response <- true
			continue
		}

		for _, e := range fe {
			wg.Add(1)
			body, err := json.Marshal(e)
			if err != nil {
				log.Printf("Couldn't marshal body to send to Spoor: %v", err)
				m.Response <- true
				continue
			}

			go func(d []byte, m queue.Message) {
				defer wg.Done()
				err := c.Send(d)
				if err != nil {
					errChan <- err
				}
			}(body, m)
		}

		go func() {
			wg.Wait()
			close(doneChan)
		}()

		select {
		case <-doneChan:
			log.Println("Sent to Spoor")
			m.Response <- true
		case err := <-errChan:
			if err != nil {
				log.Printf("Couldn't send to spoor: %v", err)
				m.Response <- false
				return
			}
		}
	}

	return nil
}
