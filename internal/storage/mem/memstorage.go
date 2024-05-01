package mem

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nikgalushko/echoevoke/internal/storage"
)

type (
	MemStorage struct {
		rw       sync.Mutex
		posts    map[string][]storage.Post
		images   map[string]image
		channels map[string]struct{}

		imageIDCounter atomic.Int64
	}

	image struct {
		data []byte
		id   int64
	}
)

func NewMemStorage() *MemStorage {
	return &MemStorage{
		posts:    make(map[string][]storage.Post),
		images:   make(map[string]image),
		channels: make(map[string]struct{}),
	}
}

func (m *MemStorage) GetLastPostID(ctx context.Context, channelID string) (int64, error) {
	m.rw.Lock()
	defer m.rw.Unlock()

	if len(m.posts[channelID]) == 0 {
		if _, ok := m.channels[channelID]; !ok {
			return 0, storage.ErrNotFound
		}

		return 0, nil
	}

	return m.posts[channelID][len(m.posts[channelID])-1].ID, nil
}

func (m *MemStorage) GetLastPost(ctx context.Context, channelID string) (storage.Post, error) {
	m.rw.Lock()
	defer m.rw.Unlock()

	if len(m.posts[channelID]) == 0 {
		return storage.Post{}, storage.ErrNotFound
	}

	return m.posts[channelID][len(m.posts[channelID])-1], nil
}

func (m *MemStorage) SavePosts(ctx context.Context, channelID string, posts []storage.Post) error {
	m.rw.Lock()
	defer m.rw.Unlock()

	m.posts[channelID] = append(m.posts[channelID], posts...)
	return nil
}

func (m *MemStorage) GetPosts(ctx context.Context, channelID string, from, to time.Time) ([]storage.Post, error) {
	m.rw.Lock()
	defer m.rw.Unlock()

	var ret []storage.Post
	for _, p := range m.posts[channelID] {
		if p.Date.After(from) && p.Date.Before(to) {
			ret = append(ret, p)
		}
	}

	return ret, nil
}

func (m *MemStorage) IsImageExists(ctx context.Context, etag string) (int64, error) {
	m.rw.Lock()
	defer m.rw.Unlock()

	img, ok := m.images[etag]
	if !ok {
		return 0, storage.ErrNotFound
	}
	return img.id, nil
}

func (m *MemStorage) SaveImage(ctx context.Context, etag string, data []byte) (int64, error) {
	m.rw.Lock()
	defer m.rw.Unlock()

	if img, ok := m.images[etag]; ok {
		return img.id, nil
	}

	img := image{data: data, id: m.imageIDCounter.Add(1)}
	m.images[etag] = img

	return img.id, nil
}

func (m *MemStorage) IsChannelRegistered(ctx context.Context, channelID string) (bool, error) {
	m.rw.Lock()
	defer m.rw.Unlock()

	_, ok := m.channels[channelID]
	return ok, nil
}

func (m *MemStorage) RegisterChannel(ctx context.Context, channelID string) error {
	m.rw.Lock()
	defer m.rw.Unlock()

	m.channels[channelID] = struct{}{}
	return nil
}

func (m *MemStorage) UnregisterChannel(ctx context.Context, channelID string) error {
	m.rw.Lock()
	defer m.rw.Unlock()

	delete(m.channels, channelID)
	return nil
}

func (m *MemStorage) AllChannels(ctx context.Context) ([]string, error) {
	m.rw.Lock()
	defer m.rw.Unlock()

	var ret []string
	for c := range m.channels {
		ret = append(ret, c)
	}
	return ret, nil
}

func (m *MemStorage) Dump(ctx context.Context, dir string) {
	m.rw.Lock()
	defer m.rw.Unlock()

	for c := range m.channels {
		rootDir := filepath.Join(dir, c)
		err := os.MkdirAll(rootDir, 0755)
		if err != nil {
			slog.Error("failed to create the channel directory", err)
			continue
		}

		for _, p := range m.posts[c] {
			err = os.WriteFile(filepath.Join(rootDir, fmt.Sprintf("%d.md", p.ID)), []byte(p.Message), 0644)
			if err != nil {
				slog.Error("failed to write the post file", err)
			}
		}

		for etag, img := range m.images {
			err = os.WriteFile(filepath.Join(rootDir, etag+".jpg"), img.data, 0644)
			if err != nil {
				slog.Error(" failed to write the image file", err)
			}
		}
	}
}
