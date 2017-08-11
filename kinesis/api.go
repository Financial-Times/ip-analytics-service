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
func PutListen(msgs <-chan queue.Message, c config.Config) err {
	creds := credentials.NewStaticCredentials(c.AWSAccessKey, c.AWSSecret, "")
	_, err := creds.Get()
	if err != nil {
		log.Printf("Couldn't get AWS credentials: %v", err)
		return
	}
	cfg := aws.NewConfig().WithRegion(c.AWSRegion).WithCredentials(creds)
	s := session.New(cfg)
	kc := kinesis.New(s)
	streamName := aws.String(c.KinesisStream)

	for m := range msgs {
		var wg sync.WaitGroup
		defer func() {
			wg.Wait()
			// TODO response on msg chan
		}()

		fe := make([]hooks.FormattedEvent, 0)
		if err := json.Unmarshal(m.Body, &fe); err != nil {
			log.Printf("Couldn't unmarshal body to put to Kinesis: %v", err)
			return
		}

		for _, e := range fe {
			wg.Add(1)
			body, err := json.Marshal(e)
			if err != nil {
				log.Printf("Couldn't marshal body to put to Kinesis: %v", err)
				return
			}

			go func(d []byte, pk string) {
				defer wg.Done()
				res, err := kc.PutRecord(&kinesis.PutRecordInput{
					Data:         d,
					StreamName:   streamName,
					PartitionKey: aws.String(pk),
				})
				if err != nil {
					log.Printf("Couldn't put record to kinesis: %v", err)
					return
				}
				log.Printf("%v\n", res)
			}(body, e.User.UUID)
		}
	}
}
