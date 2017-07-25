package hooks

import (
	"encoding/json"
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

// MembershipHandler for handling HTTP requests and publishing to queue
type MembershipHandler struct {
	Publisher
}

// HandlePOST publishes received body to queue in correct format
func (m *MembershipHandler) HandlePOST(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		errorHandler(w, "Not Found", http.StatusNotFound)
		return
	}

	_, err := parseBody(r.Body)
	if err != nil {
		errorHandler(w, err.Error(), http.StatusBadRequest)
		return
	}

	//m.Publisher.Publish(
	successHandler(w, r)
}

func parseBody(body io.Reader) (*membershipEvents, error) {
	b, err := ioutil.ReadAll(body)
	p := &membershipEvents{}
	err = json.Unmarshal(b, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

type membershipEvents struct {
	Messages []membershipEvent `json:"messages"`
}

type membershipEvent struct {
	Body               *json.RawMessage `json:"body"`
	ContentType        string           `json:"contentType"`
	MessageID          string           `json:"messageId"`
	MessageTimestamp   string           `json:"messageTimestamp"`
	MessageType        string           `json:"messageType"`
	OriginHost         string           `json:"originHost"`
	OriginHostLocation string           `json:"originHostLocation"`
	OriginSystemID     string           `json:"originSystemId"`
}

//type user struct {
//UUID string `json:"ft_guid"`
//}

// context should include a message type i.e. SubscriptionCancelRequestProcessed
//type context struct {
//}

//type system struct {
//Source  string `json:"source"`
//Version string `json:"version"`
//}

type subscriptionChange struct {
	Subscription subscription `json:"subscription"`
}

type subscription struct {
	UUID            string `json:"userId"`
	PaymentMethodID string `json:"paymentType,omitempty"`
	OfferID         string `json:"offerId,omitempty"`
	Product         struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"product,omitempty"`
	SegmentID          string `json:"segmentId,omitempty"`
	ProductRatePlanID  string
	SubscriptionID     string
	SubscriptionNumber string
	InvoiceID          string
	InvoiceNumber      string
	CancellationReason string `json:"cancellationReason,omitempty"`
}
