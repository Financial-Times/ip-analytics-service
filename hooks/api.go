package hooks

import (
	"net/http"
)

type handlerFn func(http.ResponseWriter, *http.Request)

// RegisterHandlers registers all paths and handlers to provided mux
func RegisterHandlers(mux *http.ServeMux) {
	paths := map[string]handlerFn{
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

func errorHandler(w http.ResponseWriter, msg string, status int) {
	http.Error(w, msg, status)
}
