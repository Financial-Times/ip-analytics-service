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
		return &AppError{errors.New("Method Not Allowed"), "Method Not Allowed", http.StatusMethodNotAllowed}
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

	return handleResponse(w, r, fe, m.Publish)
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
	s := system{Source: "internal-products"}
	for _, v := range me {
		log.Println("IN RANGE")
		if v.Body == "" {
			log.Println("NO BODY")
			return nil, errors.New("Bad Request - Body Required")
		}

		var err error
		var ctx interface{}
		u := user{}
		fe := FormattedEvent{}
		log.Printf("EVENT BODY: %+v", v)
		switch t := v.MessageType; t {
		case "SubscriptionPurchased", "SubscriptionCancelRequestProcessed":
			log.Printf("%+v", v)
			ctx, err = parseSubscription(&v, &u)
		case "UserProductsChanged":
			ctx, err = parseProductChange(&v, &u)
		case "UserCreated":
			ctx, err = parseUserUpdate(&v, &u)
		case "SubscriptionPaymentFailure", "SubscriptionPaymentSuccess":
			log.Printf("%+v", v)
			ctx, err = parsePayment(&v, &u)
		default:
			continue
		}
		if err != nil {
			return nil, err
		}

		// TODO make factory for formatted event
		fe.System = s
		fe.Context = ctx
		fe.User = u
		fe.Category = "membership"
		fe.Action = "change"

		log.Println("Event Info:")
		log.Printf("%+v", fe.User)
		log.Printf("%+v", fe.Context)

		e = append(e, fe)
	}
	return e, nil
}

// TODO refactor functions - remove duplication
func parseProductChange(me *membershipEvent, u *user) (*Subscription, error) {
	p := &productChange{}
	err := json.Unmarshal([]byte(me.Body), p)
	if err != nil {
		return nil, err
	}
	sub := &p.Body.Subscription
	extendSubscription(sub, me)
	extendUser(u, sub.UUID)
	return sub, nil
}

func parseSubscription(me *membershipEvent, u *user) (*Subscription, error) {
	s := &subscriptionChange{}
	err := json.Unmarshal([]byte(me.Body), s)
	if err != nil {
		return nil, err
	}
	sub := &s.Subscription
	extendSubscription(sub, me)
	extendUser(u, sub.UUID)
	return sub, nil
}

func extendSubscription(s *Subscription, m *membershipEvent) {
	s.MessageType = m.MessageType
	s.Timestamp = formatTimestamp(m.MessageTimestamp)
	s.MessageID = m.MessageID
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
	extendUser(u, upd.UUID)

	return &upd, nil
}

func parsePayment(me *membershipEvent, u *user) (*Payment, error) {
	p := &Payment{}
	err := json.Unmarshal([]byte(me.Body), p)
	if err != nil {
		return nil, err
	}
	p.MessageType = me.MessageType
	p.Timestamp = formatTimestamp(me.MessageTimestamp)
	p.MessageID = me.MessageID
	extendUser(u, p.Account.UUID)

	return p, nil
}

type subscriptionChange struct {
	Subscription Subscription `json:"subscription"`
}

type productChange struct {
	Body struct {
		Subscription Subscription `json:"user"`
	} `json:"userProductsChanged"`
}

// Subscription has necessary information for changes
type Subscription struct {
	UUID            string `json:"userId,omitempty"`
	PaymentMethodID string `json:"paymentType,omitempty"`
	OfferID         string `json:"offerId,omitempty"`
	// CHANGE
	Products           *[]Product `json:"products,omitempty"`
	SegmentID          string     `json:"segmentId,omitempty"`
	ProductRatePlanID  string     `json:"productRatePlanId,omitempty"`
	SubscriptionID     string     `json:"subscriptionId,omitempty"`
	SubscriptionNumber string     `json:"subscriptionNumber,omitempty"`
	InvoiceID          string     `json:"invoiceId,omitempty"`
	InvoiceNumber      string     `json:"invoiceNumber,omitempty"`
	CancellationReason string     `json:"cancellationReason,omitempty"`
	defaultChange
}

// Payment has payment details for failure/success
type Payment struct {
	Account struct {
		UUID string `json:"userId,omitempty"`
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"account"`
	Payment struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}
	defaultChange
}

// Product has a users subscriptions
type Product struct {
	ProductCode string `json:"productCode,omitempty"`
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
