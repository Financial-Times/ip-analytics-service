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
	{`{"uuid":"test"}`, http.StatusOK},         // valid input
	{`{uuid: test}`, http.StatusBadRequest},    // invalid JSON
	{`{"test":"test"}`, http.StatusBadRequest}, // missing UUID
}

func TestPreferencesHandlerResponse(t *testing.T) {
	var rr *httptest.ResponseRecorder
	h := &PreferenceHandler{}
	handler := http.HandlerFunc(h.HandlePOST)
	for _, tt := range responseTests {
		rr = httptest.NewRecorder()
		b := bytes.NewReader([]byte(tt.input))
		req, err := http.NewRequest("POST", "/preferences", b)
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

func TestParseBodyWithUUID(t *testing.T) {
	b := map[string]interface{}{"UUID": "1234fjf", "suppressedMarketing": true}
	bStr, err := json.Marshal(b)
	if err != nil {
		t.Errorf("JSON Marshal failed with error %v", err.Error())
	}
	pb, err := parseBody(bytes.NewReader(bStr))
	if err != nil {
		t.Errorf("parseBody returned error %v", err.Error())
	}
	if pb.UUID != b["UUID"] {
		t.Errorf("UUID failed to parse")
	}
}

func TestParseBody(t *testing.T) {
	b := &Preference{UUID: "1234fjf", SuppressedMarketing: true}
	bStr, err := json.Marshal(b)
	if err != nil {
		t.Errorf("JSON Marshal failed with error %v", err.Error())
	}
	pb, err := parseBody(bytes.NewReader(bStr))
	if err != nil {
		t.Errorf("parseBody returned error %v", err.Error())
	}
	if pb.UUID != b.UUID {
		t.Errorf("UUID failed to parse")
	}
}

func TestParseBodyWithList(t *testing.T) {
	b := &Preference{UUID: "1234fjf", SuppressedNewsletter: true, Lists: []List{List{"1234"}}}
	bStr, err := json.Marshal(b)
	if err != nil {
		t.Errorf("JSON Marshal failed with error %v", err.Error())
	}
	pb, err := parseBody(bytes.NewReader(bStr))
	if err != nil {
		t.Errorf("parseBody returned error %v", err.Error())
	}
	if pb.Lists[0] != b.Lists[0] {
		t.Errorf("Lists failed to parse")
	}
}
