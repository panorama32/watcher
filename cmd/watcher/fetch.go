package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/panorama32/watcher/internal/aggregator"
	slackclient "github.com/panorama32/watcher/internal/slack"
	"github.com/panorama32/watcher/internal/store"
	"github.com/slack-go/slack"
	"github.com/spf13/cobra"
)

func fetchMentionsCmd(client *slackclient.Client, db *store.Store) *cobra.Command {
	var count int
	var output string

	cmd := &cobra.Command{
		Use:   "mentions",
		Short: "Fetch mentions",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := client.AuthTest(); err != nil {
				return fmt.Errorf("auth test failed: %w", err)
			}

			mentions, err := client.FetchMentions(count)
			if err != nil {
				return fmt.Errorf("fetch mentions failed: %w", err)
			}

			data, err := json.MarshalIndent(mentions, "", "  ")
			if err != nil {
				return fmt.Errorf("json marshal failed: %w", err)
			}

			switch output {
			case "json":
				if err := os.WriteFile("mentions.json", data, 0644); err != nil {
					return fmt.Errorf("write file failed: %w", err)
				}
				fmt.Printf("wrote %d mentions to mentions.json\n", len(mentions))
			case "pretty":
				fmt.Println(formatMessages(mentions))
			default:
				fmt.Println(string(data))
			}
			return nil
		},
	}

	cmd.Flags().IntVarP(&count, "count", "c", 10, "number of mentions to fetch")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output format (json)")

	return cmd
}

func fetchThreadsCmd(client *slackclient.Client, db *store.Store) *cobra.Command {
	var count int
	var output string

	cmd := &cobra.Command{
		Use:   "threads",
		Short: "Fetch thread replies",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := client.AuthTest(); err != nil {
				return fmt.Errorf("auth test failed: %w", err)
			}

			threads, err := client.FetchThreadReplies(count)
			if err != nil {
				return fmt.Errorf("fetch threads failed: %w", err)
			}

			data, err := json.MarshalIndent(threads, "", "  ")
			if err != nil {
				return fmt.Errorf("json marshal failed: %w", err)
			}

			switch output {
			case "json":
				if err := os.WriteFile("threads.json", data, 0644); err != nil {
					return fmt.Errorf("write file failed: %w", err)
				}
				fmt.Printf("wrote %d threads to threads.json\n", len(threads))
			case "pretty":
				fmt.Println(formatMessages(threads))
			default:
				fmt.Println(string(data))
			}
			return nil
		},
	}

	cmd.Flags().IntVarP(&count, "count", "c", 10, "number of threads to fetch")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output format (json, pretty)")

	return cmd
}

func fetchCmd(client *slackclient.Client, db *store.Store) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch and save Slack conversations",
		RunE: func(cmd *cobra.Command, args []string) error {
			user, err := client.AuthTest()
			if err != nil {
				return fmt.Errorf("auth test failed: %w", err)
			}

			fmt.Printf("authenticated as: %s\n\n", user)

			expired, _ := db.IsUsersCacheExpired(72 * time.Hour)
			if expired {
				fmt.Println("fetching users...")
				slackUsers, err := client.FetchUsers()
				if err != nil {
					return fmt.Errorf("fetch users failed: %w", err)
				}
				storeUsers := make([]store.User, len(slackUsers))
				for i, u := range slackUsers {
					storeUsers[i] = store.User{ID: u.ID, Name: u.Name, IsBot: u.IsBot}
				}
				if err := db.ReplaceUsers(storeUsers); err != nil {
					return fmt.Errorf("save users failed: %w", err)
				}
				fmt.Printf("cached %d users\n\n", len(storeUsers))
			}

			fmt.Println("fetching mentions...")
			mentions, err := client.FetchMentions(20)
			if err != nil {
				return fmt.Errorf("fetch mentions failed: %w", err)
			}

			fmt.Println("fetching threads...")
			threads, err := client.FetchThreadReplies(20)
			if err != nil {
				return fmt.Errorf("fetch threads failed: %w", err)
			}

			messages, err := aggregator.Aggregate(mentions, threads)
			if err != nil {
				return fmt.Errorf("aggregate failed: %w", err)
			}

			if len(messages) == 0 {
				fmt.Println("no messages found")
				return nil
			}

			fmt.Println("fetching conversations...")
			convs, err := client.FetchConversations(messages)
			if err != nil {
				return fmt.Errorf("fetch conversations failed: %w", err)
			}

			fmt.Println("saving to db...")
			for _, conv := range convs {
				for _, m := range conv.Messages {
					if err := db.SaveMessage(conv.ChannelID, conv.ChannelName, m.Timestamp, m.User, m.Text); err != nil {
						fmt.Fprintf(os.Stderr, "save error: %v\n", err)
					}
				}
			}

			fmt.Printf("📋 Conversations (%d)\n\n", len(convs))
			for _, conv := range convs {
				fmt.Printf("  #%s (%d messages)\n", conv.ChannelName, len(conv.Messages))
				for _, m := range conv.Messages {
					fmt.Printf("    %s: %s\n", m.User, m.Text)
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.AddCommand(fetchMentionsCmd(client, db))
	cmd.AddCommand(fetchThreadsCmd(client, db))

	return cmd
}

func formatMessages(mentions []slack.SearchMessage) string {
	var b strings.Builder
	for i, m := range mentions {
		if i > 0 {
			b.WriteString("\n────────────────────────────────\n\n")
		}
		prefix := "#"
		if m.Channel.IsPrivate {
			prefix = "🔒"
		}
		b.WriteString(fmt.Sprintf("%s%s\n", prefix, m.Channel.Name))
		b.WriteString(fmt.Sprintf("%s: %s\n", m.Username, m.Text))
		b.WriteString(formatSlackTS(m.Timestamp))
	}
	return b.String()
}

func formatSlackTS(ts string) string {
	parts := strings.Split(ts, ".")
	if len(parts) == 0 {
		return ts
	}
	sec, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return ts
	}
	return time.Unix(sec, 0).Format("01/02 15:04")
}
