package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"

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

			userMap, err := db.LoadUserMap()
			if err != nil {
				fmt.Printf("warning: could not load user cache: %v\n", err)
				userMap = make(map[string]string)
			}
			fmt.Printf("loaded %d users from cache\n", len(userMap))

			http.HandleFunc("/conversations", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				msgs, err := db.GetConversations()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				mentionRe := regexp.MustCompile(`<@(U[A-Z0-9]+)>`)
				for i, m := range msgs {
					if name, ok := userMap[m.User]; ok {
						msgs[i].User = name
					}
					msgs[i].Text = mentionRe.ReplaceAllStringFunc(m.Text, func(match string) string {
						id := mentionRe.FindStringSubmatch(match)[1]
						if name, ok := userMap[id]; ok {
							return "@" + name
						}
						return match
					})
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
