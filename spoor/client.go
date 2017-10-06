package spoor

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/financial-times/email-news-api/newsapi"
)

// initialize global psuedo random generator for spoor region header
var regions = [2]string{"EU", "US"}

// Client is to send events to spoor
type Client struct {
	Host   string
	Client newsapi.Poster
}

// NewClient is a factory for new Clients
func NewClient(host string) *Client {
	h := &http.Client{}
	headers := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/json",
		"spoor-region": "EU", //Randomly assign a different region
		"User-Agent":   "ip-events-service/v1.1",
		"spoor-test":   "true",
	}
	c := newsapi.NewClient(headers, h)
	return &Client{host, c}
}

// Send posts the event payload to Spoor
func (c *Client) Send(body []byte) error {
	rand.Seed(time.Now().Unix())
	headers := map[string]string{
		"spoor-region": regions[rand.Intn(len(regions))],
	}

	res := make(map[string]interface{})

	_, err := c.Client.PostURL(c.Host, body, &res, headers)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("Sent to Spoor")
	return nil
}
