package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/panorama32/watcher/internal/presenter"
	"github.com/panorama32/watcher/internal/store"
	"github.com/spf13/cobra"
)

func startCmd(db *store.Store) *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the watcher server",
		RunE: func(cmd *cobra.Command, args []string) error {
			userMap, err := db.LoadUserMap()
			if err != nil {
				fmt.Printf("warning: could not load user cache: %v\n", err)
				userMap = make(map[string]string)
			}
			fmt.Printf("loaded %d users from cache\n", len(userMap))

			http.HandleFunc("/conversations", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				threads, err := db.GetConversations()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				for i, t := range threads {
					for j, m := range t.Messages {
						threads[i].Messages[j].User = presenter.ResolveUser(m.User, userMap)
						threads[i].Messages[j].Text = presenter.ResolveText(m.Text, userMap)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(threads)
			})

			addr := fmt.Sprintf(":%d", port)
			fmt.Printf("server listening on http://localhost%s\n", addr)
			return http.ListenAndServe(addr, nil)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "port to listen on")

	return cmd
}
