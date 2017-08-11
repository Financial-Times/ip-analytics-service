package kinesis

import (
	"encoding/json"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/financial-times/ip-events-service/hooks"
	"github.com/financial-times/ip-events-service/queue"
)

// PutListen starts aws/kinesis session and puts records from input chan
func PutListen(msgs <-chan queue.Message, region string, stream string) {
	creds := credentials.NewEnvCredentials()
	_, err := creds.Get()
	if err != nil {
		panic(err)
	}
	cfg := aws.NewConfig().WithRegion(region).WithCredentials(creds)
	s := session.New(cfg)
	kc := kinesis.New(s)
	streamName := aws.String(stream)

	for m := range msgs {
		fe := make([]hooks.FormattedEvent, 0)
		if err := json.Unmarshal(m.Body, &fe); err != nil {
			panic(err)
		}
		entries := make([]*kinesis.PutRecordsRequestEntry, len(fe))

		for i := 0; i < len(entries); i++ {
			d, err := json.Marshal(fe[i])
			if err != nil {
				panic(err)
			}

			entries[i] = &kinesis.PutRecordsRequestEntry{
				Data:         d,
				PartitionKey: aws.String(fe[i].User.UUID),
			}
		}

		res, err := kc.PutRecords(&kinesis.PutRecordsInput{
			Records:    entries,
			StreamName: streamName,
		})
		if err != nil {
			log.Printf("%v\n", err)
			panic(err)
		}

		log.Printf("%v\n", res)
	}
}
