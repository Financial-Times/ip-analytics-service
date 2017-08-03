package hooks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/financial-times/ip-events-service/queue"
)

var validReqTests = []struct {
	input    string
	expected int
}{
	{`{"Messages": [{"MessageType": "SubscriptionPurchased", "Body": {"uuid": "test"}}]}`, http.StatusOK},
	{`{"Messages": [{"MessageType": "SubscriptionCancelRequestProcessed", "Body": {"uuid": "test"}}]}`, http.StatusOK},
}

var pubQueue chan queue.Message

func TestMembershipHandlerOKResponse(t *testing.T) {
	pubQueue = make(chan queue.Message, 1)
	var rr *httptest.ResponseRecorder
	h := &MembershipHandler{pubQueue}
	handler := appHandler(h.HandlePOST)
	for _, tt := range validReqTests {
		rr = httptest.NewRecorder()
		b := bytes.NewReader([]byte(tt.input))
		req, err := http.NewRequest("POST", "/membership", b)
		if err != nil {
			t.Fatal(err)
		}

		// Simulate positive response from publisher
		go func() {
			msg := <-pubQueue
			msg.Response <- true
		}()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != tt.expected {
			t.Errorf("Handler returned %v for input %v but expected %v",
				status, tt.input, tt.expected)
		}
	}
}

var invalidReqTests = []struct {
	input    string
	expected int
}{
	{`In{valid{Json}`, http.StatusBadRequest},
	{`{"subscription": [{"Body": {"uuid": "test"}}]}`, http.StatusBadRequest},
	{`{"subscription": {"Messages": [{"MessageType": "Not Exist", "Body": {"uuid": "test"}}]}}`, http.StatusBadRequest},
}

func TestMembershipHandlerBadResponse(t *testing.T) {
	var rr *httptest.ResponseRecorder
	h := &MembershipHandler{pubQueue}
	handler := appHandler(h.HandlePOST)
	for _, tt := range invalidReqTests {
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
	pubQueue = make(chan queue.Message)
	b := bytes.NewReader([]byte(`[{"subscription": {"MessageType": "test"}}]`))
	rr := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/membership", b)
	if err != nil {
		t.Errorf("Problem with http request %v", err)
	}
	h := &MembershipHandler{pubQueue}
	handler := appHandler(h.HandlePOST)
	handler.ServeHTTP(rr, req)
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
