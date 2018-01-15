package keen

import (
	"github.com/financial-times/ip-events-service/hooks"
	keenSDK "github.com/keen/go-keen"
)

// Client for keen api
type Client struct {
	Client   *keenSDK.Client
	Internal bool
}

// NewClient is a factory for new clients
func NewClient(projectID string, writeKey string) *Client {
	c := &keenSDK.Client{
		ProjectID: projectID,
		WriteKey:  writeKey,
	}
	return &Client{c, true}
}

// Send FormattedEvent to keen
func (c *Client) Send(body hooks.FormattedEvent) error {
	err := c.Client.AddEvent("envoy", body)
	if err != nil {
		return err
	}
	return nil
}

// IsInternal returns true if client is meant for internal messages
func (c *Client) IsInternal() bool {
	return c.Internal
}
