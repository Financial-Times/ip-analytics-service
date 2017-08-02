package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/financial-times/ip-events-service/config"
	"github.com/financial-times/ip-events-service/hooks"
	"github.com/financial-times/ip-events-service/queue"
)

var configPath = flag.String("config", "config_dev.yaml", "path to yaml config")

func main() {
	flag.Parse()

	c, err := config.NewConfig(*configPath)
	if err != nil {
		log.Fatalln(err)
	}
	msgChan := make(chan queue.Message)
	ctx, done := context.WithCancel(context.Background())

	go func() {
		queue.Publish(queue.Redial(ctx, c.RabbitHost), msgChan, "test")
		done()
	}()

	mux := http.NewServeMux()
	hooks.RegisterHandlers(mux, c, msgChan)
	log.Printf("Server listening on %v", c.Port)
	http.ListenAndServe(":"+c.Port, mux)

	<-ctx.Done()
}
