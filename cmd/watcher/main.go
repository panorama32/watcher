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

// app holds initialized dependencies, populated by requireConfig.
type app struct {
	client *slackclient.Client
	db     *store.Store
	cfg    *config.Config
}

func main() {
	a := &app{}

	root := &cobra.Command{
		Use:   "watcher",
		Short: "Stop context-switching. Your Slack conversations, triaged and waiting.",
	}

	// configコマンドはconfig.Load()不要
	root.AddCommand(configCmd())

	// fetch/startはconfigが必要
	requireConfig := func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("config error: %w\nRun 'watcher config init' to create config file", err)
		}
		a.cfg = cfg
		a.client = slackclient.NewClient(cfg.SlackUserToken)
		dir, err := config.Dir()
		if err != nil {
			return err
		}
		a.db, err = store.New(filepath.Join(dir, "watcher.db"))
		if err != nil {
			return err
		}
		return nil
	}

	fetchCommand := fetchCmd(a)
	fetchCommand.PersistentPreRunE = requireConfig

	startCommand := startCmd(a)
	startCommand.PreRunE = requireConfig

	root.AddCommand(fetchCommand)
	root.AddCommand(startCommand)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}

	if a.db != nil {
		a.db.Close()
	}
}
