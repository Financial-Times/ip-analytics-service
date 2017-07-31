package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/financial-times/ip-events-service/config"
	"github.com/financial-times/ip-events-service/events"
	"github.com/financial-times/ip-events-service/hooks"
	//"github.com/financial-times/ip-events-service/queue"
)

var configPath = flag.String("config", "config_dev.yaml", "path to yaml config")

func main() {
	flag.Parse()

	c, err := config.NewConfig(*configPath)
	if err != nil {
		log.Fatalln(err)
	}
	ea := events.NewEventsApp(c.RabbitHost)
	if err := ea.Run(); err != nil {
		log.Printf("Something went wrong: %s", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	hooks.RegisterHandlers(mux, c)
	log.Printf("Server listening on %v", c.Port)
	http.ListenAndServe(":"+c.Port, mux)
}
