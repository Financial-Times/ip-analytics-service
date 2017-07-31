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
	Publish(body string, contentType string, ch *amqp.Channel, cfg *config.Config) error
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

	e, err := parseEvents(r.Body)
	if err != nil {
		errorHandler(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = formatEvents(e.Messages)
	if err != nil {
		errorHandler(w, err.Error(), http.StatusBadRequest)
		return
	}

	//m.Publisher.Publish(, "application/json",
	successHandler(w, r)
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

// TODO refactor all parse events to use one function and then case/type
func parseEvents(body io.Reader) (*membershipEvents, error) {
	p := &membershipEvents{}
	b, err := ioutil.ReadAll(body)
	err = json.Unmarshal(b, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// FormattedEvent published to queue for consumption
type FormattedEvent struct {
	User    user          `json:"user"`
	Context *Subscription `json:"context"`
	System  system        `json:"system"`
}

type user struct {
	UUID string `json:"ft_guid"`
}

type system struct {
	Source string `json:"source"`
}

func formatEvents(me []membershipEvent) ([]formattedEvent, error) {
	e := make([]formattedEvent, 0)
	s := system{"membership"}
	for _, v := range me {
		var err error
		var ctx *Subscription
		u := user{}
		fe := formattedEvent{}
		switch t := v.MessageType; t {
		case "SubscriptionPurchased", "SubscriptionCancelRequestProcessed":
			ctx, err = parseSubscription([]byte(*v.Body))
		default:
			return nil, errors.New("MessageType is not valid")
		}
		if err != nil {
			return nil, err
		}
		// Assign UUID to user and remove from context
		u.UUID = ctx.UUID
		ctx.UUID = ""
		ctx.MessageType = v.MessageType
		fe.System = s
		fe.Context = ctx
		fe.User = u
		e = append(e, fe)
	}
	return e, nil
}

func parseSubscription(body []byte) (*Subscription, error) {
	s := &subscriptionChange{}
	err := json.Unmarshal(body, s)
	if err != nil {
		return nil, err
	}
	return &s.Subscription, nil
}

type subscriptionChange struct {
	Subscription Subscription `json:"subscription"`
}

// Subscription has necessary information for changes
type Subscription struct {
	UUID            string `json:"userId,omitempty"`
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
	MessageType        string `json:"messageType"`
}
