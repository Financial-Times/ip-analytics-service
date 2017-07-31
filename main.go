package main

import (
	"flag"
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
		log.Println(err)
		panic(err)
	}

	conn, connErr, err := queue.DialRabbit(c.RabbitHost)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	log.Println(conn, connErr)
}
