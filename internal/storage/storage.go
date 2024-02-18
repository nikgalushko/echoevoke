package storage

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

type (
	Post struct {
		ID      int64
		Date    time.Time
		Message string
		Type    string
		Images  []byte
	}

	Storage interface {
		SavePost(channelID string, post Post) error
		GetPosts(channelID string, from, to time.Time) ([]Post, error)
		GetLastPost(channelID string) (Post, error)
		GetLastPostID(channelID string) (int64, error)
	}
)
