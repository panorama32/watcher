package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "watcher",
		Short: "Stop context-switching. Your Slack conversations, triaged and waiting.",
	}

	root.AddCommand(fetchCmd())
	root.AddCommand(startCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
