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

// PreferenceHandler for handling HTTP requests and publishing to queue
type PreferenceHandler struct {
	Publish chan queue.Message
}

// HandlePOST publishes received body to queue in correct format
func (m *PreferenceHandler) HandlePOST(w http.ResponseWriter, r *http.Request) *AppError {
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

	e, err := parsePreferenceEvent(reader)
	if err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("json error at byte offset %d", e.Offset)
		}
		return &AppError{err, "Bad Request", http.StatusBadRequest}
	}

	fe, err := formatPreferenceEvent(e)
	if err != nil {
		return &AppError{err, "Bad Request", http.StatusBadRequest}
	}

	if fe == nil {
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

type preferenceEvent struct {
	Body             string `json:"body"`
	ContentType      string `json:"contentType"`
	MessageID        string `json:"messageId"`
	MessageTimestamp string `json:"messageTimestamp"`
	MessageType      string `json:"messageType"`
}

// TODO refactor all parse events to use one function and then case/type
func parsePreferenceEvent(body io.ReadCloser) (*preferenceEvent, error) {
	p := &preferenceEvent{}
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func formatPreferenceEvent(p *preferenceEvent) ([]FormattedEvent, error) {
	e := make([]FormattedEvent, 0)
	s := system{"internal-products"}
	var err error
	var ctx *preference
	u := user{}
	fe := FormattedEvent{}

	switch t := p.MessageType; t {
	case "UserPreferenceUpdated", "UserPreferenceCreated":
		ctx, err = parsePreference([]byte(p.Body))
	default:
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// Assign UUID to user and remove from context
	u.UUID = ctx.UUID
	ctx.MessageType = p.MessageType
	ctx.Timestamp = formatTimestamp(p.MessageTimestamp)
	ctx.MessageID = p.MessageID
	fe.System = s
	fe.Context = ctx
	fe.User = u
	e = append(e, fe)

	return e, nil
}

func parsePreference(body []byte) (*preference, error) {
	s := &preference{}
	err := json.Unmarshal(body, s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Preference has necessary information for changes
type preference struct {
	UUID                     string   `json:"uuid,omitempty"`
	SuppressedMarketing      bool     `json:"suppressedMarketing"`
	SuppressedNewsletter     bool     `json:"suppressedNewsletter"`
	SuppressedRecommendation bool     `json:"suppressedRecommendation"`
	SuppressedAccount        bool     `json:"suppressedAccount"`
	Expired                  bool     `json:"expired"`
	Lists                    []string `json:"lists"`
	ModifiedPaths            []string `json:"modifiedPaths"`
	MessageType              string   `json:"messageType"`
	MessageID                string   `json:"messageId"`
	Timestamp                string   `json:"timestamp"`
}
