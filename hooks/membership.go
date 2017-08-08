package hooks

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/financial-times/ip-events-service/queue"
)

// MembershipHandler for handling HTTP requests and publishing to queue
type MembershipHandler struct {
	Publish chan queue.Message
}

// HandlePOST publishes received body to queue in correct format
func (m *MembershipHandler) HandlePOST(w http.ResponseWriter, r *http.Request) *AppError {
	if r.Method != "POST" {
		return &AppError{errors.New("Not Found"), "Not Found", http.StatusNotFound}
	}

	e, err := parseEvents(r.Body)
	if err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("json error at byte offset %d", e.Offset)
		}
		return &AppError{err, "Bad Request", http.StatusBadRequest}
	}

	fe, err := formatEvents(e.Messages)
	if err != nil {
		return &AppError{err, "Bad Request", http.StatusBadRequest}
	}

	body, err := json.Marshal(fe)
	if err != nil {
		return &AppError{err, "Bad Request", http.StatusBadRequest}
	}

	confirm := make(chan bool, 1) // Create a confirm channel to wait for confirmation from publisher
	msg := queue.Message{
		Body:     body,
		Response: confirm,
	}
	m.Publish <- msg

	ok := <-confirm
	if !ok {
		return &AppError{errors.New("Internal Server Error"), "Internal Server Error", http.StatusInternalServerError}
	}

	successHandler(w, r)
	return nil
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
	log.Printf("Body: %v", b)
	err = json.Unmarshal(b, p)
	if err != nil {
		return nil, err
	}
	if len(p.Messages) == 0 {
		return nil, errors.New("No valid message events")
	}

	return p, nil
}

func formatEvents(me []membershipEvent) ([]FormattedEvent, error) {
	e := make([]FormattedEvent, 0)
	s := system{"membership"}
	for _, v := range me {
		if v.Body == nil {
			return nil, errors.New("Bad Request - Body Required")
		}

		var err error
		var ctx *Subscription
		u := user{}
		fe := FormattedEvent{}
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

// Subscription has necessary information for changes
type Subscription struct {
	UUID            string `json:"userId"`
	PaymentMethodID string `json:"paymentType,omitempty"`
	OfferID         string `json:"offerId,omitempty"`
	Product         struct {
		ID   string `json:"id,omitempty"`
		Name string `json:"name,omitempty"`
	} `json:"product,omitempty"`
	SegmentID          string `json:"segmentId,omitempty"`
	ProductRatePlanID  string `json:"productRatePlanId,omitempty"`
	SubscriptionID     string `json:"subscriptionId,omitempty"`
	SubscriptionNumber string `json:"subscriptionNumber,omitempty"`
	InvoiceID          string `json:"invoiceId,omitempty"`
	InvoiceNumber      string `json:"invoiceNumber,omitempty"`
	CancellationReason string `json:"cancellationReason,omitempty"`
	MessageType        string `json:"messageType"`
}
