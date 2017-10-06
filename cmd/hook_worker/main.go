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
	"github.com/financial-times/ip-events-service/spoor"
)

var configPath = flag.String("config", "config_dev.yaml", "path to yaml config")

func main() {
	flag.Parse()
	c, err := config.NewConfig(*configPath)
	if err != nil {
		log.Fatalln(err)
	}

	var writer io.Writer
	var msgChan chan queue.Message

	ctx, done := context.WithCancel(context.Background())

	switch c.GOENV {
	case "production":
		msgChan = make(chan queue.Message)
		go func() {
			cl := spoor.NewClient(c.SpoorHost)
			spoor.Consume(msgChan, cl)
			done()
		}()
	case "staging":
		writer = ioutil.Discard
		msgChan = queue.Write(writer)
	default:
		msgChan = queue.Write(os.Stdout)
	}

	go func() {
		queue.Consume(queue.Redial(ctx, c.RabbitHost, c.QueueName), msgChan, c.QueueName)
		done()
	}()
	<-ctx.Done()
}
