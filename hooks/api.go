package hooks

import (
	"encoding/json"
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
		prefix + "/membership":       &MembershipHandler{publish},
		prefix + "/user-preferences": &PreferenceHandler{publish},
		prefix + "/marketing":        &MarketingHandler{publish},
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

func handleResponse(w http.ResponseWriter, r *http.Request, fe []FormattedEvent, pub chan queue.Message) *AppError {
	if len(fe) == 0 {
		successHandler(w, r)
		return nil
	}

	b, err := json.Marshal(fe)
	if err != nil {
		return &AppError{err, "Bad Request", http.StatusBadRequest}
	}

	confirm := make(chan bool, 1)
	msg := queue.Message{
		Body:     b,
		Response: confirm,
	}
	pub <- msg

	ok := <-confirm
	if !ok {
		return &AppError{errors.New("Internal Server Error"), "Internal Server Error", http.StatusInternalServerError}
	}

	successHandler(w, r)
	return nil
}

func extendUser(u *user, uuid string) {
	u.UUID = uuid
	u.EnrichmentUUID = uuid
}

// FormattedEvent published to queue for consumption
type FormattedEvent struct {
	User     user        `json:"user"`
	Context  interface{} `json:"context"`
	Category string      `json:"category"`
	Action   string      `json:"action"`
	System   system      `json:"system"`
}

type baseEvent struct {
	Body             string `json:"body"`
	ContentType      string `json:"contentType"`
	MessageID        string `json:"messageId"`
	MessageTimestamp string `json:"messageTimestamp"`
	MessageType      string `json:"messageType"`
}

type user struct {
	UUID           string `json:"ft_guid"`
	EnrichmentUUID string `json:"uuid"`
}

type system struct {
	Source string `json:"source"`
}

type defaultChange struct {
	MessageType string `json:"messageType"`
	MessageID   string `json:"messageId"`
	Timestamp   string `json:"timestamp"`
}
