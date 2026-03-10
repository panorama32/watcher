package main

import (
	"fmt"
	"os"

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

	fmt.Printf("authenticated as: %s\n", user)
}
