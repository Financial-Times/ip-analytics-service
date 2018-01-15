package hooks

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/financial-times/ip-events-service/queue"
)

// MarketingHandler for handling HTTP requests and publishing to queue
type MarketingHandler struct {
	Publish chan queue.Message
}

// HandlePOST publishes received body to queue in correct format
func (m *MarketingHandler) HandlePOST(w http.ResponseWriter, r *http.Request) *AppError {
	if r.Method != "POST" {
		return &AppError{errors.New("Method Not Allowed"), "Method Not Allowed", http.StatusMethodNotAllowed}
	}

	e, err := parseMarketingEvents(r.Body)
	if err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("json error at byte offset %d", e.Offset)
		}
		return &AppError{err, "Bad Request", http.StatusBadRequest}
	}

	fe, err := formatMarketingEvents(e.Events)
	if err != nil {
		return &AppError{err, "Bad Request", http.StatusBadRequest}
	}
	return handleResponse(w, r, fe, m.Publish)
}

type marketingEvents struct {
	Events []baseEvent `json:"events"`
}

func parseMarketingEvents(body io.ReadCloser) (*marketingEvents, error) {
	m := &marketingEvents{}
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, m)
	if err != nil {
		return nil, err
	}
	if len(m.Events) == 0 {
		return nil, errors.New("No valid message events")
	}

	return m, nil
}

func formatMarketingEvents(me []baseEvent) ([]FormattedEvent, error) {
	e := make([]FormattedEvent, 0)
	s := system{Source: "internal-products"}
	for _, v := range me {
		if v.Body == "" {
			return nil, errors.New("Bad Request - Body Required")
		}

		var err error
		var ctx *EntityProgression
		u := user{}
		fe := FormattedEvent{}
		switch t := v.MessageType; t {
		case "EntityProgressed":
			ctx, err = parseEntityProgression([]byte(v.Body))
		default:
			continue
		}
		if err != nil {
			return nil, err
		}

		if strings.ToLower(ctx.EntityType) == "user" {
			extendUser(&u, ctx.UUID)
		}

		ctx.MessageType = v.MessageType
		ctx.Timestamp = formatTimestamp(v.MessageTimestamp)
		ctx.MessageID = v.MessageID
		fe.System = s
		fe.Context = ctx
		fe.User = u
		fe.Category = "marketing-automation"
		fe.Action = "progression"
		fe.Internal = true
		e = append(e, fe)
	}
	return e, nil
}

// EntityProgression contains details of enytity track progression (silo to silo)
type EntityProgression struct {
	UUID          string `json:"entityId"`
	EntityType    string `json:"entityType"`
	TrackID       int    `json:"trackId"`
	TrackRevID    int    `json:"trackRevId"`
	OriginSiloID  int    `json:"originSiloId"`
	LandingSiloID int    `json:"landingSiloId"`
	RuleSetID     int    `json:"ruleSetId"`
	defaultChange
}

func parseEntityProgression(body []byte) (*EntityProgression, error) {
	ep := &EntityProgression{}
	err := json.Unmarshal(body, ep)
	if err != nil {
		return nil, err
	}
	return ep, nil
}
