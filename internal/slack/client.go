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
		Sort:          "timestamp",
		SortDirection: "desc",
		Count:         20,
	}

	msgs, err := c.api.SearchMessages(query, params)
	if err != nil {
		return nil, err
	}

	return msgs.Matches, nil
}

func (c *Client) FetchThreadReplies() ([]slack.SearchMessage, error) {
	if c.userID == "" {
		return nil, fmt.Errorf("userID is not set; call AuthTest first")
	}

	query := fmt.Sprintf("from:<@%s> is:thread", c.userID)
	params := slack.SearchParameters{
		Sort:          "timestamp",
		SortDirection: "desc",
		Count:         20,
	}

	msgs, err := c.api.SearchMessages(query, params)
	if err != nil {
		return nil, err
	}

	return msgs.Matches, nil
}

type Conversation struct {
	ChannelID   string
	ChannelName string
	Messages    []slack.Message
}

func (c *Client) FetchConversations(searchMessages []slack.SearchMessage) ([]Conversation, error) {
	var convs []Conversation

	for _, sm := range searchMessages {
		msgs, _, _, err := c.api.GetConversationReplies(&slack.GetConversationRepliesParameters{
			ChannelID: sm.Channel.ID,
			Timestamp: sm.Timestamp,
		})
		if err != nil {
			// スレッドではない単体メッセージの場合
			convs = append(convs, Conversation{
				ChannelID:   sm.Channel.ID,
				ChannelName: sm.Channel.Name,
				Messages: []slack.Message{{
					Msg: slack.Msg{Text: sm.Text, Timestamp: sm.Timestamp, User: sm.User},
				}},
			})
			continue
		}

		convs = append(convs, Conversation{
			ChannelID:   sm.Channel.ID,
			ChannelName: sm.Channel.Name,
			Messages:    msgs,
		})
	}

	return convs, nil
}
