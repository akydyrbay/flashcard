package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	"flashcard/storage"
)

type Storage struct {
	db *sql.DB
}

func New(path string) (*Storage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("can't open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("can't connect to database: %w", err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) Init(ctx context.Context) error {
	q := `CREATE TABLE IF NOT EXISTS items (
        hash TEXT PRIMARY KEY,
        user_name TEXT,
        name TEXT,
        content TEXT
    )`
	if _, err := s.db.ExecContext(ctx, q); err != nil {
		return fmt.Errorf("can't create table: %w", err)
	}
	return nil
}

func (s *Storage) Save(ctx context.Context, it *storage.Item) error {
	h, err := it.Hash()
	if err != nil {
		return err
	}
	q := `INSERT INTO items (hash, user_name, name, content) VALUES (?, ?, ?, ?)`
	if _, err := s.db.ExecContext(ctx, q, h, it.UserName, it.Name, it.Content); err != nil {
		return fmt.Errorf("can't save item: %w", err)
	}
	return nil
}

func (s *Storage) Get(ctx context.Context, userName, name string) (*storage.Item, error) {
	q := `SELECT content FROM items WHERE user_name = ? AND name = ? LIMIT 1`
	var content string
	err := s.db.QueryRowContext(ctx, q, userName, name).Scan(&content)
	if err == sql.ErrNoRows {
		return nil, storage.ErrNoSavedItems
	}
	if err != nil {
		return nil, fmt.Errorf("can't get item: %w", err)
	}
	return &storage.Item{UserName: userName, Name: name, Content: content}, nil
}

func (s *Storage) IsExists(ctx context.Context, it *storage.Item) (bool, error) {
	q := `SELECT COUNT(*) FROM items WHERE user_name = ? AND name = ?`
	var count int
	if err := s.db.QueryRowContext(ctx, q, it.UserName, it.Name).Scan(&count); err != nil {
		return false, fmt.Errorf("can't check if item exists: %w", err)
	}
	return count > 0, nil
}

func (s *Storage) Remove(ctx context.Context, it *storage.Item) error {
	q := `DELETE FROM items WHERE user_name = ? AND name = ?`
	if _, err := s.db.ExecContext(ctx, q, it.UserName, it.Name); err != nil {
		return fmt.Errorf("can't remove item: %w", err)
	}
	return nil
}

func (s *Storage) List(ctx context.Context, userName string) ([]string, error) {
	q := `SELECT name FROM items WHERE user_name = ?`
	rows, err := s.db.QueryContext(ctx, q, userName)
	if err != nil {
		return nil, fmt.Errorf("can't list items: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("can't scan name: %w", err)
		}
		names = append(names, name)
	}
	return names, nil
}
