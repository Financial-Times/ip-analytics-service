package hooks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/financial-times/ip-events-service/config"
	"github.com/streadway/amqp"
)

type mockPublisher struct {
	mockPublish func(body string, contentType string, ch *amqp.Channel, cfg *config.Config) error
}

func (mp *mockPublisher) Publish(body string, contentType string, ch *amqp.Channel, cfg *config.Config) error {
	if mp.mockPublish != nil {
		return mp.mockPublish("", "", nil, nil)
	}
	return nil
}

var responseTests = []struct {
	input    string
	expected int
}{
	{`{"uuid":"test"}`, http.StatusOK},       // valid input
	{`{uuid: test}}`, http.StatusBadRequest}, // invalid JSON
}

func TestMembershipHandlerResponse(t *testing.T) {
	var rr *httptest.ResponseRecorder
	h := &MembershipHandler{}
	handler := http.HandlerFunc(h.HandlePOST)
	for _, tt := range responseTests {
		rr = httptest.NewRecorder()
		b := bytes.NewReader([]byte(tt.input))
		req, err := http.NewRequest("POST", "/membership", b)
		if err != nil {
			t.Fatal(err)
		}

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != tt.expected {
			t.Errorf("Handler returned %v for input %v but expected %v",
				status, tt.input, tt.expected)
		}
	}
}

func TestHandlePublishEvents(t *testing.T) {
	called := false
	b := bytes.NewReader([]byte(`[{"subscription": {"MessageType": "test"}}]`))
	rr := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/membership", b)
	if err != nil {
		t.Errorf("Problem with http request %v", err)
	}
	h := &MembershipHandler{
		&mockPublisher{
			mockPublish: func(body string, contentType string, ch *amqp.Channel, cfg *config.Config) error {
				called = true
				return nil
			},
		},
	}
	handler := http.HandlerFunc(h.HandlePOST)
	handler.ServeHTTP(rr, req)

	if !called {
		t.Error("Expected publisher to be called")
	}
}

func TestParseEvents(t *testing.T) {
	msg := json.RawMessage(`{"UUID": "123"}`)
	b := &membershipEvents{[]membershipEvent{membershipEvent{Body: &msg}}}
	bStr, err := json.Marshal(b)
	if err != nil {
		t.Errorf("JSON Marshal failed with error %v", err.Error())
	}
	pb, err := parseEvents(bytes.NewReader(bStr))
	if err != nil {
		t.Errorf("parseEvents returned error %v", err.Error())
	}
	if *pb.Messages[0].Body == nil {
		t.Errorf("Expected parsed message to equal %v but got %v", b.Messages[0], pb.Messages[0])
	}
}

func TestFormatEvents(t *testing.T) {
	msg := json.RawMessage(`{"subscription": {"userId": "123"}}`)
	m := []membershipEvent{membershipEvent{MessageType: "SubscriptionPurchased", Body: &msg}}
	f, err := formatEvents(m)
	if err != nil {
		t.Errorf("Could not format events due to error %v", err)
	}
	for _, v := range f {
		if v.User.UUID == "" {
			t.Error("Expected UUID on formatted event User")
		}
		if v.Context.MessageType == "" {
			t.Error("Expected MessageType on formatted event Context")
		}
	}
}
