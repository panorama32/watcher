package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/panorama32/watcher/internal/config"
	slackclient "github.com/panorama32/watcher/internal/slack"
	"github.com/panorama32/watcher/internal/store"
	"github.com/spf13/cobra"
)

func main() {
	// load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	// create slack client
	client := slackclient.NewClient(cfg.SlackUserToken)

	dir, err := config.Dir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config dir error: %v\n", err)
		os.Exit(1)
	}

	db, err := store.New(filepath.Join(dir, "watcher.db"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "store error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	root := &cobra.Command{
		Use:   "watcher",
		Short: "Stop context-switching. Your Slack conversations, triaged and waiting.",
	}

	root.AddCommand(fetchCmd(client, db))
	root.AddCommand(startCmd(client, db, cfg))

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
