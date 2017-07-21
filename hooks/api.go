package hooks

import (
	"net/http"
)

type HandlerFn func(http.ResponseWriter, *http.Request)

func RegisterHandlers(mux *http.ServeMux) {
	paths := map[string]HandlerFn{
		"/hello": preferencesHandler,
	}
	for p, h := range paths {
		mux.HandleFunc(p, h)
	}
}

func successHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("OK"))
}
