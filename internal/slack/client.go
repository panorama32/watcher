package slack

import (
	"fmt"
	"net/url"
	"strings"

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

func (c *Client) UserID() string {
	return c.userID
}

func (c *Client) AuthTest() (string, error) {
	res, err := c.api.AuthTest()
	if err != nil {
		return "", err
	}
	c.userID = res.UserID
	return res.User, nil
}

func (c *Client) FetchMentions(count int) ([]slack.SearchMessage, error) {
	if c.userID == "" {
		return nil, fmt.Errorf("userID is required")
	}
	if count < 1 {
		return nil, fmt.Errorf("count must be at least 1 (got %d)", count)
	}
	if count > 100 {
		return nil, fmt.Errorf("count must be 100 or less (got %d); pagination is not supported", count)
	}

	query := fmt.Sprintf("<@%s>", c.userID)
	params := slack.SearchParameters{
		Sort:          "timestamp",
		SortDirection: "desc",
		Count:         count,
	}

	msgs, err := c.api.SearchMessages(query, params)
	if err != nil {
		return nil, err
	}

	return msgs.Matches, nil
}

func (c *Client) FetchThreadReplies(count int) ([]slack.SearchMessage, error) {
	if c.userID == "" {
		return nil, fmt.Errorf("userID is not set; call AuthTest first")
	}

	query := fmt.Sprintf("from:<@%s> is:thread", c.userID)
	params := slack.SearchParameters{
		Sort:          "timestamp",
		SortDirection: "desc",
		Count:         count,
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

func (c *Client) FetchConversation(permalink string) (Conversation, error) {
	u, err := url.Parse(permalink)
	if err != nil {
		return Conversation{}, fmt.Errorf("failed to parse permalink %q: %w", permalink, err)
	}

	// /archives/{channel_id}/{p<ts>}
	parts := strings.Split(u.Path, "/")
	if len(parts) < 4 {
		return Conversation{}, fmt.Errorf("unexpected permalink format: %q", permalink)
	}
	channelID := parts[2]

	ts, err := threadTSFromPermalink(permalink)
	if err != nil {
		return Conversation{}, err
	}

	msgs, _, _, err := c.api.GetConversationReplies(&slack.GetConversationRepliesParameters{
		ChannelID: channelID,
		Timestamp: ts,
	})
	if err != nil {
		return Conversation{}, fmt.Errorf("get conversation replies: %w", err)
	}

	return Conversation{
		ChannelID: channelID,
		Messages:  msgs,
	}, nil
}

// threadTSFromPermalink extracts the thread root timestamp from a permalink.
// For replies: uses thread_ts query param.
// For thread roots: extracts from path /archives/{channel}/{p<ts>} and converts to dot format.
func threadTSFromPermalink(permalink string) (string, error) {
	u, err := url.Parse(permalink)
	if err != nil {
		return "", fmt.Errorf("failed to parse permalink %q: %w", permalink, err)
	}

	if threadTS := u.Query().Get("thread_ts"); threadTS != "" {
		return threadTS, nil
	}

	// /archives/{channel_id}/{p<sec><micro>} → "<sec>.<micro>"
	parts := strings.Split(u.Path, "/")
	if len(parts) >= 4 && len(parts[3]) > 1 && parts[3][0] == 'p' {
		digits := parts[3][1:]
		if len(digits) > 6 {
			return digits[:len(digits)-6] + "." + digits[len(digits)-6:], nil
		}
	}

	return "", fmt.Errorf("unexpected permalink format: %q", permalink)
}

func (c *Client) FetchConversations(searchMessages []slack.SearchMessage) ([]Conversation, error) {
	var convs []Conversation

	for _, sm := range searchMessages {
		ts, err := threadTSFromPermalink(sm.Permalink)
		if err != nil {
			return nil, fmt.Errorf("failed to extract thread_ts: %w", err)
		}

		msgs, _, _, err := c.api.GetConversationReplies(&slack.GetConversationRepliesParameters{
			ChannelID: sm.Channel.ID,
			Timestamp: ts,
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

type User struct {
	ID    string
	Name  string
	IsBot bool
}

func (c *Client) FetchUsers() ([]User, error) {
	var users []User

	slackUsers, err := c.api.GetUsers()
	if err != nil {
		return nil, err
	}

	for _, u := range slackUsers {
		if u.Deleted {
			continue
		}
		name := u.Profile.DisplayName
		if name == "" {
			name = u.RealName
		}
		users = append(users, User{ID: u.ID, Name: name, IsBot: u.IsBot})
	}

	return users, nil
}
