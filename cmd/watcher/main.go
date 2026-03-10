package main

import (
	"fmt"
	"os"

	"github.com/panorama32/watcher/internal/aggregator"
	"github.com/panorama32/watcher/internal/config"
	slackclient "github.com/panorama32/watcher/internal/slack"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	client := slackclient.NewClient(cfg.SlackUserToken)
	user, err := client.AuthTest()
	if err != nil {
		fmt.Fprintf(os.Stderr, "auth test failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("authenticated as: %s\n\n", user)

	mentions, err := client.FetchMentions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch mentions failed: %v\n", err)
		os.Exit(1)
	}

	threads, err := client.FetchThreadReplies()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch threads failed: %v\n", err)
		os.Exit(1)
	}

	messages := aggregator.Aggregate(mentions, threads)

	if len(messages) == 0 {
		fmt.Println("no messages found")
		return
	}

	convs, err := client.FetchConversations(messages)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch conversations failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📋 Conversations (%d)\n\n", len(convs))
	for _, conv := range convs {
		fmt.Printf("  #%s (%d messages)\n", conv.ChannelName, len(conv.Messages))
		for _, m := range conv.Messages {
			fmt.Printf("    %s: %s\n", m.User, m.Text)
		}
		fmt.Println()
	}
}
