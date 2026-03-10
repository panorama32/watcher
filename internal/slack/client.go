package slack

import (
	"fmt"

	"github.com/slack-go/slack"
)

type Client struct {
	api    *slack.Client
	userID string
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
	c.userID = res.UserID
	return res.User, nil
}

func (c *Client) FetchMentions() ([]slack.SearchMessage, error) {
	if c.userID == "" {
		return nil, fmt.Errorf("userID is not set; call AuthTest first")
	}

	query := fmt.Sprintf("<@%s>", c.userID)
	params := slack.SearchParameters{
		Sort:  "timestamp",
		SortDirection: "desc",
		Count: 20,
	}

	msgs, err := c.api.SearchMessages(query, params)
	if err != nil {
		return nil, err
	}

	return msgs.Matches, nil
}
