package storage

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

type (
	Post struct {
		ID int64
		// Date is the date and time of the post
		Date time.Time
		// Message is the text of the post in markdown format
		Message string
		// Images is the list of images etag that are in the post
		Images []string
	}

	// PostsStorage stores the posts
	PostsStorage interface {
		SavePosts(channelID string, post []Post) error
		GetPosts(channelID string, from, to time.Time) ([]Post, error)
		GetLastPost(channelID string) (Post, error)
		GetLastPostID(channelID string) (int64, error)
	}

	// ImagesStorage stores the images blobs
	ImagesStorage interface {
		IsImageExists(etag string) (bool, error)
		SaveImage(etag string, data []byte) error
	}

	// ChannelRegistry stores the channels that are registered to be scraped
	ChannelsRegistry interface {
		IsChannelRegistered(channelID string) (bool, error)
		RegisterChannel(channelID string) error
		UnregisterChannel(channelID string) error
	}
)
