package spoor

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"
  "fmt"

	"github.com/financial-times/email-news-api/newsapi"
	"github.com/financial-times/ip-events-service/hooks"
)

// initialize global psuedo random generator for spoor region header
var regions = [2]string{"EU", "US"}

// Client is to send events to spoor
type Client struct {
	Host     string
	Client   newsapi.Poster
	Internal bool
}

// NewClient is a factory for new Clients
func NewClient(host string) *Client {
	timeout := time.Duration(5 * time.Second)
	h := &http.Client{
		Timeout: timeout,
	}
	headers := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/json",
		"spoor-region": "EU", //Randomly assign a different region
		"User-Agent":   "ip-events-service/v1.1",
		//"spoor-test":   "true",
	}
	c := newsapi.NewClient(headers, h)
	return &Client{host, c, false}
}

// Send posts the event payload to Spoor
func (c *Client) Send(body hooks.FormattedEvent) error {
	rand.Seed(time.Now().Unix())
	headers := map[string]string{
		"spoor-region": regions[rand.Intn(len(regions))],
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil
	}
	res := make(map[string]interface{})

  fmt.Printf("JSON %s", b)

	_, err = c.Client.PostURL(c.Host, b, &res, headers)
	if err != nil {
		return err
	}
	return nil
}

// IsInternal returns true if client is meant for internal messages
func (c *Client) IsInternal() bool {
	return c.Internal
}
