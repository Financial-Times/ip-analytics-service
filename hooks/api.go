package hooks

import (
	"log"
	"net/http"
)

type HandlerFn func(http.ResponseWriter, *http.Request)

func RegisterHandlers(mux *http.ServeMux) {
	paths := map[string]HandlerFn{
		"/hello", helloHandler,
	}
	for p, h := range paths {
		mux.HandleFunc(p, h)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write("Hello, World!")
}
