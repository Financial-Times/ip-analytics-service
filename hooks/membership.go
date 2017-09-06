package hooks

import (
	"compress/gzip"
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

	var reader io.ReadCloser
	var err error
	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(r.Body)
		if err != nil {
			return &AppError{err, "Bad Request", http.StatusBadRequest}
		}
		defer reader.Close()
	default:
		reader = r.Body
	}

	e, err := parseEvents(reader)
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

	if len(fe) == 0 {
		successHandler(w, r)
		return nil
	}

	b, err := json.Marshal(fe)
	if err != nil {
		return &AppError{err, "Bad Request", http.StatusBadRequest}
	}

	confirm := make(chan bool, 1) // Create a confirm channel to wait for confirmation from publisher
	msg := queue.Message{
		Body:     b,
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
	Body               string `json:"body"`
	ContentType        string `json:"contentType"`
	MessageID          string `json:"messageId"`
	MessageTimestamp   string `json:"messageTimestamp"`
	MessageType        string `json:"messageType"`
	OriginHost         string `json:"originHost"`
	OriginHostLocation string `json:"originHostLocation"`
	OriginSystemID     string `json:"originSystemId"`
}

// TODO refactor all parse events to use one function and then case/type
func parseEvents(body io.ReadCloser) (*membershipEvents, error) {
	p := &membershipEvents{}
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
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
	s := system{"internal-products"}
	for _, v := range me {
		if v.Body == "" {
			return nil, errors.New("Bad Request - Body Required")
		}

		var err error
		var ctx interface{}
		u := user{}
		fe := FormattedEvent{}
		switch t := v.MessageType; t {
		case "SubscriptionPurchased", "SubscriptionCancelRequestProcessed":
			ctx, err = parseSubscription(&v, &u)
		case "UserCreated":
			ctx, err = parseUserUpdate(&v, &u)
		default:
			continue
		}
		if err != nil {
			return nil, err
		}

		fe.System = s
		fe.Context = ctx
		fe.User = u
		e = append(e, fe)
	}
	return e, nil
}

// TODO refactor functions - remove duplication
func parseSubscription(me *membershipEvent, u *user) (*Subscription, error) {
	s := &subscriptionChange{}
	err := json.Unmarshal([]byte(me.Body), s)
	if err != nil {
		return nil, err
	}
	sub := s.Subscription
	sub.MessageType = me.MessageType
	sub.Timestamp = formatTimestamp(me.MessageTimestamp)
	sub.MessageID = me.MessageID
	u.UUID = sub.UUID
	return &sub, nil
}

func parseUserUpdate(me *membershipEvent, u *user) (*Update, error) {
	up := &UserUpdate{}
	err := json.Unmarshal([]byte(me.Body), up)
	if err != nil {
		return nil, err
	}
	upd := up.Update
	upd.MessageType = me.MessageType
	upd.Timestamp = formatTimestamp(me.MessageTimestamp)
	upd.MessageID = me.MessageID
	u.UUID = upd.UUID
	return &upd, nil
}

type subscriptionChange struct {
	Subscription Subscription `json:"subscription"`
}

type defaultChange struct {
	MessageType string `json:"messageType"`
	MessageID   string `json:"messageId"`
	Timestamp   string `json:"timestamp"`
}

// Subscription has necessary information for changes
type Subscription struct {
	UUID            string `json:"userId,omitempty"`
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
	defaultChange
}

// UserUpdate represents a new or updated user
type UserUpdate struct {
	Update Update `json:"user"`
}

// Update details of UserUpdate
type Update struct {
	UUID string `json:"id,omitempty"`
	defaultChange
	// private fields
	email     string
	firstName string
	lastName  string
}
