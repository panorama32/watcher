package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/panorama32/watcher/internal/config"
	"github.com/panorama32/watcher/internal/store"
	"github.com/spf13/cobra"
)

func startCmd() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the watcher server",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := config.Dir()
			if err != nil {
				return fmt.Errorf("config dir error: %w", err)
			}

			db, err := store.New(filepath.Join(dir, "watcher.db"))
			if err != nil {
				return fmt.Errorf("store error: %w", err)
			}
			defer db.Close()

			http.HandleFunc("/conversations", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				msgs, err := db.GetConversations()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(msgs)
			})

			addr := fmt.Sprintf(":%d", port)
			fmt.Printf("server listening on http://localhost%s\n", addr)
			return http.ListenAndServe(addr, nil)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "port to listen on")

	return cmd
}
