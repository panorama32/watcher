package store

import (
	"database/sql"

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

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS conversations (
			channel_id   TEXT NOT NULL,
			channel_name TEXT NOT NULL,
			ts           TEXT NOT NULL,
			user         TEXT NOT NULL,
			text         TEXT NOT NULL,
			PRIMARY KEY (channel_id, ts)
		)
	`)
	return err
}
