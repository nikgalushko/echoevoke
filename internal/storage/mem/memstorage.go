package mem

import (
	"sync"
	"time"

	"github.com/nikgalushko/echoevoke/internal/storage"
)

type MemStorage struct {
	rw     sync.Mutex
	posts  map[string][]storage.Post
	images map[string][]byte
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		posts:  make(map[string][]storage.Post),
		images: make(map[string][]byte),
	}
}

func (m *MemStorage) GetLastPostID(channelID string) (int64, error) {
	m.rw.Lock()
	defer m.rw.Unlock()

	if len(m.posts[channelID]) == 0 {
		return 0, storage.ErrNotFound
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
