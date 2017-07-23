package main

import (
	"net/http"

	"github.com/financial-times/ip-events-service/hooks"
)

func main() {
	mux := http.NewServeMux()
	hooks.RegisterHandlers(mux)
	http.ListenAndServe(":8000", mux)
}