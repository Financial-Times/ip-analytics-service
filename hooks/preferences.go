package hooks

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/financial-times/ip-events-service/config"
	"github.com/streadway/amqp"
)

// Publisher interface for rabbitmq publisher
type Publisher interface {
	Publish(body string, contentType string, ch *amqp.Channel, cfg config.Config) error
}

// PreferenceHandler for handling HTTP requests and publishing to queue
type PreferenceHandler struct {
	Publisher
}

// HandlePOST publishes received body to queue in correct format
func (ph *PreferenceHandler) HandlePOST(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		errorHandler(w, "Not Found", http.StatusNotFound)
		return
	}

	_, err := parseBody(r.Body)
	if err != nil {
		errorHandler(w, err.Error(), http.StatusBadRequest)
		return
	}

	successHandler(w, r)
}

func parseBody(body io.Reader) (*Preference, error) {
	b, err := ioutil.ReadAll(body)
	p := &Preference{}
	err = json.Unmarshal(b, p)
	if err != nil {
		return nil, err
	}
	if p.UUID == "" {
		return nil, errors.New("Missing mandatory field: UUID")
	}
	return p, nil
}

// Preference defines the possible changes to user preferences - keyed by UUID
type Preference struct {
	UUID                     string `json:"uuid"`
	SuppressedMarketing      string `json:"suppressedMarketing"`
	SuppressedNewsletter     string `json:"suppressedNewsletter"`
	SuppressedRecommendation string `json:"suppressedRecommendation"`
	SuppressedAccount        string `json:"suppressedAccount"`
	SubscriptionChange       List   `json:"list"`
}

// List contains an ID of an email list and the action (ADD/DEL)
type List struct {
	ListID string
	Action string
}
