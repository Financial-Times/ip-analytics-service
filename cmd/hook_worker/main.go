package main

import (
	"context"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/financial-times/ip-events-service/config"
	"github.com/financial-times/ip-events-service/kinesis"
	"github.com/financial-times/ip-events-service/queue"
	"github.com/financial-times/ip-events-service/spoor"
)

var configPath = flag.String("config", "config_dev.yaml", "path to yaml config")

func main() {
	flag.Parse()
	conf, err := config.NewConfig(*configPath)
	if err != nil {
		log.Fatalln(err)
	}

	var writer io.Writer
	var msgChan chan queue.Message

	ctx, done := context.WithCancel(context.Background())

	switch conf.GOENV {
	case "production":
		msgChan = make(chan queue.Message)
		spoorChan := make(chan queue.Message)
		kinesisChan := make(chan queue.Message)
		go func() {
			cl := spoor.NewClient(conf.SpoorHost)
			spoor.Consume(spoorChan, cl)
			done()
		}()
		go func() {
			kinesis.PutListen(kinesisChan, conf)
			done()
		}()
		go func() {
			for {
				msg := <-msgChan
				kinesisChan <- msg
				spoorChan <- msg
			}
		}()
	case "staging":
		writer = ioutil.Discard
		msgChan = queue.Write(writer)
	default:
		msgChan = queue.Write(os.Stdout)
	}

	go func() {
		queue.Consume(queue.Redial(ctx, conf.RabbitHost, conf.QueueName), msgChan, conf.QueueName)
		done()
	}()
	<-ctx.Done()
}
