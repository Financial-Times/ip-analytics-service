package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/financial-times/ip-events-service/config"
	"github.com/financial-times/ip-events-service/hooks"
)

var configPath = flag.String("config", "config_dev.yaml", "path to yaml config")

func main() {
	flag.Parse()

	c, err := config.NewConfig(*configPath)
	if err != nil {
		log.Fatalln(err)
	}
	mux := http.NewServeMux()
	hooks.RegisterHandlers(mux, c)
	log.Println("Server listening on 8000")
	http.ListenAndServe(":8000", mux)
}
