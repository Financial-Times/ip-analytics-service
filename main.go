package main

import (
	"flag"
	"log"

	"github.com/financial-times/ip-analytics-service/config"
	"github.com/financial-times/ip-analytics-service/queue"
)

var configPath = flag.String("config", "config_dev.yaml", "path to yaml config")

func main() {
	flag.Parse()

	c, err := config.NewConfig(*configPath)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	conn, connErr, err := queue.DialRabbit()
	if err != nil {
		log.Println(err)
		panic(err)
	}
}
