package hooks

import (
	"log"
	"net/http"

	"github.com/financial-times/ip-events-service/config"
	"github.com/financial-times/ip-events-service/queue"
)

// Handler handles webhook events from a particular service
type Handler interface {
	HandlePOST(w http.ResponseWriter, r *http.Request)
}

// RegisterHandlers registers all paths and handlers to provided mux
func RegisterHandlers(mux *http.ServeMux, cfg config.Config, publish chan queue.Message) {
	prefix := "/webhooks"
	paths := map[string]Handler{
		prefix + "/membership": &MembershipHandler{publish},
	}
	for p, h := range paths {
		mux.Handle(p, authMiddleware(http.HandlerFunc(h.HandlePOST), cfg.APIKey))
	}
}

func authMiddleware(f http.HandlerFunc, key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-KEY") != key {
			errorHandler(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		f(w, r)
	}
}

func successHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("OK"))
}

type reqError struct {
	Error   error
	Message string
	Status  int
}

func errorHandler(w http.ResponseWriter, e reqError) {
	log.Printf("%v", err)
	http.Error(w, e.Message, e.Status)
}
