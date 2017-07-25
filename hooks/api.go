package hooks

import (
	"github.com/financial-times/ip-events-service/config"
	"net/http"
)

// Handler handles webhook events from a particular service
type Handler interface {
	HandlePOST(w http.ResponseWriter, r *http.Request)
}

// RegisterHandlers registers all paths and handlers to provided mux
func RegisterHandlers(mux *http.ServeMux, cfg config.Config) {
	paths := map[string]Handler{
		"/membership": &MembershipHandler{},
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

func errorHandler(w http.ResponseWriter, msg string, status int) {
	http.Error(w, msg, status)
}
