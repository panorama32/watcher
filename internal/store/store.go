package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

type Thread struct {
	ChannelID   string    `json:"channel_id"`
	ChannelName string    `json:"channel_name"`
	ThreadTS    string    `json:"thread_ts"`
	Messages    []Message `json:"messages"`
}

type Message struct {
	Ts   string `json:"ts"`
	User string `json:"user"`
	Text string `json:"text"`
}

func (s *Store) SaveConversation(channelID, channelName, threadTS string, messages []Message) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		INSERT OR REPLACE INTO threads (channel_id, channel_name, thread_ts)
		VALUES (?, ?, ?)
	`, channelID, channelName, threadTS); err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO messages (channel_id, thread_ts, ts, user, text)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, m := range messages {
		if _, err := stmt.Exec(channelID, threadTS, m.Ts, m.User, m.Text); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) GetConversations() ([]Thread, error) {
	rows, err := s.db.Query(`
		SELECT t.channel_id, t.channel_name, t.thread_ts, m.ts, m.user, m.text
		FROM threads t
		JOIN messages m ON t.channel_id = m.channel_id AND t.thread_ts = m.thread_ts
		JOIN (
			SELECT channel_id, thread_ts, MAX(ts) AS last_ts
			FROM messages
			GROUP BY channel_id, thread_ts
		) latest ON t.channel_id = latest.channel_id AND t.thread_ts = latest.thread_ts
		ORDER BY latest.last_ts DESC, m.ts ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	threadMap := make(map[string]*Thread)
	var order []string

	for rows.Next() {
		var channelID, channelName, threadTS, ts, user, text string
		if err := rows.Scan(&channelID, &channelName, &threadTS, &ts, &user, &text); err != nil {
			return nil, err
		}
		key := channelID + ":" + threadTS
		if _, ok := threadMap[key]; !ok {
			threadMap[key] = &Thread{
				ChannelID:   channelID,
				ChannelName: channelName,
				ThreadTS:    threadTS,
			}
			order = append(order, key)
		}
		threadMap[key].Messages = append(threadMap[key].Messages, Message{Ts: ts, User: user, Text: text})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	threads := make([]Thread, 0, len(order))
	for _, key := range order {
		threads = append(threads, *threadMap[key])
	}
	return threads, nil
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	IsBot bool   `json:"is_bot"`
}

func (s *Store) IsUsersCacheExpired(ttl time.Duration) (bool, error) {
	var fetchedAt string
	err := s.db.QueryRow(`SELECT fetched_at FROM users LIMIT 1`).Scan(&fetchedAt)
	if err == sql.ErrNoRows {
		return true, nil
	}
	if err != nil {
		return true, nil
	}

	t, err := time.Parse(time.RFC3339, fetchedAt)
	if err != nil {
		return true, nil
	}

	return time.Since(t) > ttl, nil
}

func (s *Store) ReplaceUsers(users []User) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM users`); err != nil {
		return err
	}

	stmt, err := tx.Prepare(`INSERT INTO users (id, name, is_bot, fetched_at) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	start := time.Now()
	for _, u := range users {
		if _, err := stmt.Exec(u.ID, u.Name, u.IsBot, now); err != nil {
			return err
		}
	}
	fmt.Printf("inserted %d users in %.1fs\n", len(users), time.Since(start).Seconds())

	return tx.Commit()
}

func (s *Store) GetUserName(id string) (string, error) {
	var name string
	err := s.db.QueryRow(`SELECT name FROM users WHERE id = ?`, id).Scan(&name)
	if err != nil {
		return id, err
	}
	return name, nil
}

func (s *Store) LoadUserMap() (map[string]string, error) {
	rows, err := s.db.Query(`SELECT id, name FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[string]string)
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		m[id] = name
	}
	return m, rows.Err()
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS threads (
			channel_id   TEXT NOT NULL,
			channel_name TEXT NOT NULL,
			thread_ts    TEXT NOT NULL,
			PRIMARY KEY (channel_id, thread_ts)
		);
		CREATE TABLE IF NOT EXISTS messages (
			channel_id TEXT NOT NULL,
			thread_ts  TEXT NOT NULL,
			ts         TEXT NOT NULL,
			user       TEXT NOT NULL,
			text       TEXT NOT NULL,
			PRIMARY KEY (channel_id, ts),
			FOREIGN KEY (channel_id, thread_ts) REFERENCES threads(channel_id, thread_ts)
		);
		CREATE TABLE IF NOT EXISTS users (
			id         TEXT PRIMARY KEY,
			name       TEXT NOT NULL,
			is_bot     INTEGER NOT NULL DEFAULT 0,
			fetched_at TEXT NOT NULL
		);
	`)
	return err
}
