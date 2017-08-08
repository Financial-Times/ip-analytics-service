package main

import (
	"context"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/financial-times/ip-events-service/config"
	"github.com/financial-times/ip-events-service/queue"
)

var configPath = flag.String("config", "config_dev.yaml", "path to yaml config")

func main() {
	flag.Parse()
	c, err := config.NewConfig(*configPath)
	if err != nil {
		log.Fatalln(err)
	}
	var writer io.Writer
	if c.GOENV == "production" {
		writer = ioutil.Discard
	} else {
		writer = os.Stdout
	}
	msgChan := queue.Write(writer)
	ctx, done := context.WithCancel(context.Background())

	go func() {
		queue.Consume(queue.Redial(ctx, c.RabbitHost, c.QueueName), msgChan, c.QueueName)
		done()
	}()

	<-ctx.Done()
}
