package hooks

import (
	"errors"
	"log"
	"net/http"

	"github.com/financial-times/ip-events-service/config"
	"github.com/financial-times/ip-events-service/queue"
)

// AppError combines error with HTTP response details
type AppError struct {
	Error   error
	Message string
	Status  int
}

type appHandler func(http.ResponseWriter, *http.Request) *AppError

// Handler handles webhook events from a particular service
type Handler interface {
	HandlePOST(http.ResponseWriter, *http.Request) *AppError
}

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil { // e is *appError, not os.Error
		log.Printf("%v", e.Error)
		http.Error(w, e.Message, e.Status)
	}
}

// RegisterHandlers registers all paths and handlers to provided mux
func RegisterHandlers(mux *http.ServeMux, cfg config.Config, publish chan queue.Message) {
	prefix := "/webhooks"
	paths := map[string]Handler{
		prefix + "/membership": &MembershipHandler{publish},
	}
	for p, h := range paths {
		mux.Handle(p, authMiddleware(h.HandlePOST, cfg.APIKey))
	}
}

func authMiddleware(f appHandler, key string) appHandler {
	return func(w http.ResponseWriter, r *http.Request) *AppError {
		if r.Header.Get("X-API-KEY") != key {
			return &AppError{errors.New("Unauthorized"), "Unauthorized", http.StatusUnauthorized}
		}
		return f(w, r)
	}
}

func successHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("OK"))
}
