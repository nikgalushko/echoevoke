package mem

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/nikgalushko/echoevoke/internal/storage"
)

type MemStorage struct {
	rw       sync.Mutex
	posts    map[string][]storage.Post
	images   map[string][]byte
	channels map[string]struct{}
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		posts:    make(map[string][]storage.Post),
		images:   make(map[string][]byte),
		channels: make(map[string]struct{}),
	}
}

func (m *MemStorage) GetLastPostID(channelID string) (int64, error) {
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

func (m *MemStorage) GetLastPost(channelID string) (storage.Post, error) {
	m.rw.Lock()
	defer m.rw.Unlock()

	if len(m.posts[channelID]) == 0 {
		return storage.Post{}, storage.ErrNotFound
	}

	return m.posts[channelID][len(m.posts[channelID])-1], nil
}

func (m *MemStorage) SavePosts(channelID string, posts []storage.Post) error {
	m.rw.Lock()
	defer m.rw.Unlock()

	m.posts[channelID] = append(m.posts[channelID], posts...)
	return nil
}

func (m *MemStorage) GetPosts(channelID string, from, to time.Time) ([]storage.Post, error) {
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

func (m *MemStorage) IsImageExists(etag string) (bool, error) {
	m.rw.Lock()
	defer m.rw.Unlock()

	_, ok := m.images[etag]
	return ok, nil
}

func (m *MemStorage) SaveImage(etag string, data []byte) error {
	m.rw.Lock()
	defer m.rw.Unlock()

	m.images[etag] = data
	return nil
}

func (m *MemStorage) IsChannelRegistered(channelID string) (bool, error) {
	m.rw.Lock()
	defer m.rw.Unlock()

	_, ok := m.channels[channelID]
	return ok, nil
}

func (m *MemStorage) RegisterChannel(channelID string) error {
	m.rw.Lock()
	defer m.rw.Unlock()

	m.channels[channelID] = struct{}{}
	return nil
}

func (m *MemStorage) UnregisterChannel(channelID string) error {
	m.rw.Lock()
	defer m.rw.Unlock()

	delete(m.channels, channelID)
	return nil
}

func (m *MemStorage) Dump(dir string) {
	m.rw.Lock()
	defer m.rw.Unlock()

	for c := range m.channels {
		rootDir := filepath.Join(dir, c)
		err := os.MkdirAll(rootDir, 0755)
		if err != nil {
			log.Println("[ERROR] failed to create the channel directory", err)
			continue
		}

		for _, p := range m.posts[c] {
			err = os.WriteFile(filepath.Join(rootDir, fmt.Sprintf("%d.md", p.ID)), []byte(p.Message), 0644)
			if err != nil {
				log.Println("[ERROR] failed to write the post file", err)
			}
		}

		for etag, blob := range m.images {
			err = os.WriteFile(filepath.Join(rootDir, etag+".jpg"), blob, 0644)
			if err != nil {
				log.Println("[ERROR] failed to write the image file", err)
			}
		}
	}
}
