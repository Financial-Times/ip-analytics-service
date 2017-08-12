package hooks

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/financial-times/ip-events-service/queue"
)

const msgType = "UserPreferenceUpdated"

func TestPreferenceHandlerOKResponse(t *testing.T) {
	var validReqTests = []struct {
		input    string
		expected int
	}{
		{`{"MessageType": "` + msgType + `","Body": "{\"uuid\": \"test\"}"}`, http.StatusOK},
	}

	pubQueue = make(chan queue.Message, 1)
	var rr *httptest.ResponseRecorder
	h := &PreferenceHandler{pubQueue}
	handler := appHandler(h.HandlePOST)
	for _, tt := range validReqTests {
		rr = httptest.NewRecorder()
		b := bytes.NewReader([]byte(tt.input))
		req, err := http.NewRequest("POST", "/webhooks/user-preferences", b)
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

func TestPreferenceHandlerFalseConfirm(t *testing.T) {
	pubQueue = make(chan queue.Message, 1)
	rr := httptest.NewRecorder()
	h := &PreferenceHandler{pubQueue}
	handler := appHandler(h.HandlePOST)
	msg := `{"MessageType": "` + msgType + `", "Body": "{\"uuid\": \"test\"}"}`
	b := bytes.NewReader([]byte(msg))
	req, err := http.NewRequest("POST", "/webhooks/user-preferences", b)
	if err != nil {
		t.Fatal(err)
	}

	// Simulate positive response from publisher
	go func() {
		msg := <-pubQueue
		msg.Response <- false
	}()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Handler returned %v but expected %v",
			status, http.StatusInternalServerError)
	}
}

func TestPreferenceHandlerBadResponse(t *testing.T) {
	var invalidReqTests = []struct {
		input    string
		expected int
	}{
		{`In{valid{Json}`, http.StatusBadRequest},                                         // bad JSON
		{`{"MessageType": "Not Exist", "Body": {"uuid": "test"}}`, http.StatusBadRequest}, // Non-existant MessageType
		{`{"MessageType": "` + msgType + `"}]}}`, http.StatusBadRequest},                  // Missing Body
	}

	var rr *httptest.ResponseRecorder
	h := &PreferenceHandler{pubQueue}
	handler := appHandler(h.HandlePOST)
	for _, tt := range invalidReqTests {
		rr = httptest.NewRecorder()
		b := bytes.NewReader([]byte(tt.input))
		req, err := http.NewRequest("POST", "/webhooks/user-preferences", b)
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

func TestHandlePublishPreferenceEvent(t *testing.T) {
	pubQueue = make(chan queue.Message)
	b := bytes.NewReader([]byte(`[{"subscription": {"MessageType": "test"}}]`))
	rr := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/webhooks/user-preferences", b)
	if err != nil {
		t.Errorf("Problem with http request %v", err)
	}
	h := &PreferenceHandler{pubQueue}
	handler := appHandler(h.HandlePOST)
	handler.ServeHTTP(rr, req)
}

func TestParsePreferenceEvent(t *testing.T) {
	msg := `{"UUID": "123"}`
	b := &preferenceEvent{Body: msg}
	bStr, err := json.Marshal(b)
	if err != nil {
		t.Errorf("JSON Marshal failed with error %v", err.Error())
	}
	pb, err := parsePreferenceEvent(ioutil.NopCloser(bytes.NewReader(bStr)))
	if err != nil {
		t.Errorf("parseEvents returned error %v", err.Error())
	}
	if pb.Body == "" {
		t.Errorf("Expected parsed message to equal %v but got %v", b, pb)
	}
}

func TestFormatPreference(t *testing.T) {
	msg := `{"uuid": "123"}`
	m := &preferenceEvent{MessageType: msgType, Body: msg}
	f, err := formatPreferenceEvent(m)
	if err != nil {
		t.Errorf("Could not format events due to error %v", err)
	}
	for _, v := range f {
		if v.User.UUID == "" {
			t.Error("Expected UUID on formatted event User")
		}
		if _, ok := v.Context.(*preference); !ok {
			t.Error("Expected Context to be a pointer to Update")
		}
	}
}
