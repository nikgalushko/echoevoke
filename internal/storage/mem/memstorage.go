package mem

import (
	"time"

	"github.com/nikgalushko/echoevoke/internal/storage"
)

type MemStorage struct {
	posts map[string][]storage.Post
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		posts: make(map[string][]storage.Post),
	}
}

func (m *MemStorage) GetLastPostID(channelID string) (int64, error) {
	if len(m.posts[channelID]) == 0 {
		return 0, storage.ErrNotFound
	}

	return m.posts[channelID][len(m.posts[channelID])-1].ID, nil
}

func (m *MemStorage) GetLastPost(channelID string) (storage.Post, error) {
	if len(m.posts[channelID]) == 0 {
		return storage.Post{}, storage.ErrNotFound
	}

	return m.posts[channelID][len(m.posts[channelID])-1], nil
}

func (m *MemStorage) SavePost(channelID string, post storage.Post) error {
	m.posts[channelID] = append(m.posts[channelID], post)
	return nil
}

func (m *MemStorage) GetPosts(channelID string, from, to time.Time) ([]storage.Post, error) {
	var ret []storage.Post
	for _, p := range m.posts[channelID] {
		if p.Date.After(from) && p.Date.Before(to) {
			ret = append(ret, p)
		}
	}

	return ret, nil
}
