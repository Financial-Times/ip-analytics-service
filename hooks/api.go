package hooks

import (
	"net/http"
)

// Handler handles webhook events from a particular service
type Handler interface {
	HandlePOST
}

// RegisterHandlers registers all paths and handlers to provided mux
func RegisterHandlers(mux *http.ServeMux) {
	paths := map[string]Handler{
		"/hello": &PreferenceHandler{},
	}
	for p, h := range paths {
		mux.Handle(p, http.HandlerFunc(h.HandlePOST))
	}
}

func successHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("OK"))
}

func errorHandler(w http.ResponseWriter, msg string, status int) {
	http.Error(w, msg, status)
}
