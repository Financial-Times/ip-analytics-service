package hooks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

func TestParseBody(t *testing.T) {
	msg := json.RawMessage(`{"UUID": "123"}`)
	b := &membershipEvents{[]membershipEvent{membershipEvent{Body: &msg}}}
	bStr, err := json.Marshal(b)
	if err != nil {
		t.Errorf("JSON Marshal failed with error %v", err.Error())
	}
	pb, err := parseBody(bytes.NewReader(bStr))
	if err != nil {
		t.Errorf("parseBody returned error %v", err.Error())
	}
	if *pb.Messages[0].Body == nil {
		t.Errorf("Expected parsed message to equal %v but got %v", b.Messages[0], pb.Messages[0])
	}
}

func TestFormatEvent(t *testing.T) {
}
