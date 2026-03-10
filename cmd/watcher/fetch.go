package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/panorama32/watcher/internal/aggregator"
	"github.com/panorama32/watcher/internal/config"
	slackclient "github.com/panorama32/watcher/internal/slack"
	"github.com/panorama32/watcher/internal/store"
	"github.com/spf13/cobra"
)

func fetchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fetch",
		Short: "Fetch and save Slack conversations",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("config error: %w", err)
			}

			client := slackclient.NewClient(cfg.SlackUserToken)
			user, err := client.AuthTest()
			if err != nil {
				return fmt.Errorf("auth test failed: %w", err)
			}

			fmt.Printf("authenticated as: %s\n\n", user)

			dir, err := config.Dir()
			if err != nil {
				return fmt.Errorf("config dir error: %w", err)
			}

			db, err := store.New(filepath.Join(dir, "watcher.db"))
			if err != nil {
				return fmt.Errorf("store error: %w", err)
			}
			defer db.Close()

			mentions, err := client.FetchMentions()
			if err != nil {
				return fmt.Errorf("fetch mentions failed: %w", err)
			}

			threads, err := client.FetchThreadReplies()
			if err != nil {
				return fmt.Errorf("fetch threads failed: %w", err)
			}

			messages := aggregator.Aggregate(mentions, threads)

			if len(messages) == 0 {
				fmt.Println("no messages found")
				return nil
			}

			convs, err := client.FetchConversations(messages)
			if err != nil {
				return fmt.Errorf("fetch conversations failed: %w", err)
			}

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
}
