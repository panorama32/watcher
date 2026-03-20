package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/panorama32/watcher/internal/presenter"
	"github.com/spf13/cobra"
)

func startCmd(a *app) *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the watcher server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := a.client.AuthTest(); err != nil {
				return fmt.Errorf("auth test failed: %w", err)
			}

			userMap, err := a.db.LoadUserMap()
			if err != nil {
				fmt.Printf("warning: could not load user cache: %v\n", err)
				userMap = make(map[string]string)
			}
			fmt.Printf("loaded %d users from cache\n", len(userMap))

			// initial fetch
			fmt.Println("running initial fetch...")
			if err := runFetch(a); err != nil {
				fmt.Printf("initial fetch error: %v\n", err)
			}

			// periodic fetch
			interval, err := time.ParseDuration(a.cfg.FetchInterval)
			if err != nil || interval <= 0 {
				interval = 3 * time.Minute
				fmt.Printf("using default fetch interval: %s\n", interval)
			} else {
				fmt.Printf("fetch interval: %s\n", interval)
			}

			go func() {
				ticker := time.NewTicker(interval)
				defer ticker.Stop()
				for range ticker.C {
					fmt.Printf("[%s] fetching...\n", time.Now().Format("15:04:05"))
					if err := runFetch(a); err != nil {
						fmt.Printf("fetch error: %v\n", err)
					}
					// refresh user map after fetch
					if updated, err := a.db.LoadUserMap(); err == nil {
						userMap = updated
					}
				}
			}()

			http.HandleFunc("/conversations", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				threads, err := a.db.GetConversations()
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
