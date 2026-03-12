package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/panorama32/watcher/internal/aggregator"
	slackclient "github.com/panorama32/watcher/internal/slack"
	"github.com/panorama32/watcher/internal/store"
	"github.com/spf13/cobra"
)

func fetchMentionsCmd(client *slackclient.Client, db *store.Store) *cobra.Command {
	var count int

	cmd := &cobra.Command{
		Use:   "mentions",
		Short: "Fetch mentions",
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := client.AuthTest()
			if err != nil {
				return fmt.Errorf("auth test failed: %w", err)
			}

			mentions, err := client.FetchMentions(userID, count)
			if err != nil {
				return fmt.Errorf("fetch mentions failed: %w", err)
			}

			data, err := json.MarshalIndent(mentions, "", "  ")
			if err != nil {
				return fmt.Errorf("json marshal failed: %w", err)
			}

			if err := os.WriteFile("mentions.json", data, 0644); err != nil {
				return fmt.Errorf("write file failed: %w", err)
			}

			fmt.Printf("wrote %d mentions to mentions.json\n", len(mentions))
			return nil
		},
	}

	cmd.Flags().IntVarP(&count, "count", "c", 20, "number of mentions to fetch")

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
			mentions, err := client.FetchMentions(client.UserID(), 20)
			if err != nil {
				return fmt.Errorf("fetch mentions failed: %w", err)
			}

			fmt.Println("fetching threads...")
			threads, err := client.FetchThreadReplies()
			if err != nil {
				return fmt.Errorf("fetch threads failed: %w", err)
			}

			messages := aggregator.Aggregate(mentions, threads)

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

	return cmd
}
