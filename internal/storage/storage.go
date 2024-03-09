package storage

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

type (
	Post struct {
		ID      int64
		Date    time.Time // Date is the date and time of the post
		Message string    // Message is the text of the post in markdown format
		Images  []int64   // Images is the list of images id that are in the post
	}

	// PostsStorage stores the posts
	PostsStorage interface {
		SavePosts(ctx context.Context, channelID string, post []Post) error
		GetPosts(ctx context.Context, channelID string, from, to time.Time) ([]Post, error)
		GetLastPost(ctx context.Context, channelID string) (Post, error)
		GetLastPostID(ctx context.Context, channelID string) (int64, error)
	}

	// ImagesStorage stores the images blobs
	ImagesStorage interface {
		IsImageExists(ctx context.Context, etag string) (int64, error)
		SaveImage(ctx context.Context, etag string, data []byte) (int64, error)
	}

	// ChannelRegistry stores the channels that are registered to be scraped
	ChannelsRegistry interface {
		IsChannelRegistered(ctx context.Context, channelID string) (bool, error)
		RegisterChannel(ctx context.Context, channelID string) error
		UnregisterChannel(ctx context.Context, channelID string) error
		AllChannels(ctx context.Context) ([]string, error)
	}
)
