package hooks

import (
	"bytes"
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

func TestPreferencesHandlerPublish(t *testing.T) {

}
