package kinesis

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/financial-times/ip-events-service/config"
	"github.com/financial-times/ip-events-service/hooks"
	"github.com/financial-times/ip-events-service/queue"
)

// PutListen starts aws/kinesis session and puts records from input chan
func PutListen(msgs <-chan queue.Message, c config.Config) error {
	creds := credentials.NewStaticCredentials(c.AWSAccessKey, c.AWSSecret, "")
	_, err := creds.Get()
	if err != nil {
		log.Printf("Couldn't get AWS credentials: %v", err)
		return err
	}
	cfg := aws.NewConfig().WithRegion(c.AWSRegion).WithCredentials(creds)
	s := session.New(cfg)
	kc := kinesis.New(s)
	streamName := aws.String(c.KinesisStream)

	for m := range msgs {
		var wg sync.WaitGroup
		errChan := make(chan error, 1)
		doneChan := make(chan bool, 1)

		fe := make([]hooks.FormattedEvent, 0)
		if err := json.Unmarshal(m.Body, &fe); err != nil {
			log.Printf("Couldn't unmarshal body to put to Kinesis: %v", err)
			continue
		}

		for _, e := range fe {
			// Anonymous user - don't partition on uuid kinesis stream
			if e.User.UUID == "" {
				continue
			}
			wg.Add(1)
			body, err := json.Marshal(e)
			if err != nil {
				log.Printf("Couldn't marshal body to put to Kinesis: %v", err)
				continue
			}

			go func(d []byte, pk string) {
				defer wg.Done()
				_, err := kc.PutRecord(&kinesis.PutRecordInput{
					Data:         d,
					StreamName:   streamName,
					PartitionKey: aws.String(pk),
				})
				if err != nil {
					errChan <- err
				} else {
					log.Println("Sent to Kinesis")
				}
			}(body, e.User.UUID)
		}

		go func() {
			wg.Wait()
			close(doneChan)
			// TODO response on msg chan
		}()

		select {
		case <-doneChan:
		case err := <-errChan:

			if err != nil {
				log.Printf("Couldn't put record to kinesis: %v", err)
			}
		}
	}

	return nil
}
