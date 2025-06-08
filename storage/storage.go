package storage

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"

	"flashcard/lib/e"
)

// ErrNoSavedItems indicates no items found for a user
var ErrNoSavedItems = errors.New("no saved items")

// Storage defines put/get/remove by name per user
type Storage interface {
	Save(ctx context.Context, it *Item) error
	Get(ctx context.Context, user, name string) (*Item, error)
	Remove(ctx context.Context, it *Item) error
	IsExists(ctx context.Context, it *Item) (bool, error)
	List(ctx context.Context, user string) ([]string, error)
}

// Item represents a named text entry by a user
type Item struct {
	Name     string
	Content  string
	UserName string
}

func (i Item) Hash() (string, error) {
	h := sha1.New()
	if _, err := io.WriteString(h, i.UserName); err != nil {
		return "", e.Wrap("can't calculate hash", err)
	}
	if _, err := io.WriteString(h, i.Name); err != nil {
		return "", e.Wrap("can't calculate hash", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
