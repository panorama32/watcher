package slack

import (
	"github.com/slack-go/slack"
)

type Client struct {
	api *slack.Client
}

func NewClient(userToken string) *Client {
	return &Client{
		api: slack.New(userToken),
	}
}

func (c *Client) AuthTest() (string, error) {
	res, err := c.api.AuthTest()
	if err != nil {
		return "", err
	}
	return res.User, nil
}
