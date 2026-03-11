package store

import (
	"database/sql"
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

func (s *Store) SaveMessage(channelID, channelName, ts, user, text string) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO conversations (channel_id, channel_name, ts, user, text)
		VALUES (?, ?, ?, ?, ?)
	`, channelID, channelName, ts, user, text)
	return err
}

type Message struct {
	ChannelID   string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	Ts          string `json:"ts"`
	User        string `json:"user"`
	Text        string `json:"text"`
}

func (s *Store) GetConversations() ([]Message, error) {
	rows, err := s.db.Query(`SELECT channel_id, channel_name, ts, user, text FROM conversations ORDER BY ts DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ChannelID, &m.ChannelName, &m.Ts, &m.User, &m.Text); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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

	now := time.Now().UTC().Format(time.RFC3339)
	for _, u := range users {
		if _, err := tx.Exec(`INSERT INTO users (id, name, fetched_at) VALUES (?, ?, ?)`, u.ID, u.Name, now); err != nil {
			return err
		}
	}

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
		CREATE TABLE IF NOT EXISTS conversations (
			channel_id   TEXT NOT NULL,
			channel_name TEXT NOT NULL,
			ts           TEXT NOT NULL,
			user         TEXT NOT NULL,
			text         TEXT NOT NULL,
			PRIMARY KEY (channel_id, ts)
		);
		CREATE TABLE IF NOT EXISTS users (
			id         TEXT PRIMARY KEY,
			name       TEXT NOT NULL,
			fetched_at TEXT NOT NULL
		);
	`)
	return err
}
