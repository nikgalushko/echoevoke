package disk

import (
	"errors"
	"testing"
	"time"

	"github.com/matryer/is"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nikgalushko/echoevoke/internal/storage"
)

func TestPostsStorage(t *testing.T) {
	toTime := func(unix int64) time.Time { return time.Unix(unix, 0).UTC() }
	s := NewPostsStorage(db)

	t.Run("posts exist", func(t *testing.T) {
		const channelWithPosts = "channel1"
		is := is.New(t)

		posts := []storage.Post{
			{ID: 1, Date: toTime(123), Message: "message1", Images: []int64{1, 2}},
			{ID: 2, Date: toTime(124), Message: "message2"},
			{ID: 3, Date: toTime(125), Message: "message3", Images: []int64{2, 3}},
			{ID: 4, Date: toTime(126), Images: []int64{4}},
		}
		err := s.SavePosts(channelWithPosts, posts)
		is.NoErr(err)

		t.Run("get posts where one has two images and another not", func(t *testing.T) {
			is := is.New(t)

			actualPosts, err := s.GetPosts(channelWithPosts, toTime(124), toTime(126))
			is.NoErr(err)
			is.Equal(len(actualPosts), 2)
			is.Equal(posts[1:3], actualPosts)
		})

		t.Run("last post id", func(t *testing.T) {
			is := is.New(t)

			lastPostID, err := s.GetLastPostID(channelWithPosts)
			is.NoErr(err)
			is.Equal(lastPostID, posts[len(posts)-1].ID)
		})

		t.Run("get posts where one has no images", func(t *testing.T) {
			is := is.New(t)

			actualPosts, err := s.GetPosts(channelWithPosts, toTime(124), toTime(125))
			is.NoErr(err)
			is.Equal(len(actualPosts), 1)
			is.Equal(posts[1], actualPosts[0])
		})

		t.Run("get posts where all posts have images", func(t *testing.T) {
			is := is.New(t)

			actualPosts, err := s.GetPosts(channelWithPosts, toTime(123), toTime(127))
			is.NoErr(err)
			is.Equal(len(actualPosts), len(posts))
			is.Equal(posts, actualPosts)
		})
	})

	t.Run("posts not exist", func(t *testing.T) {
		const channelWithoutPosts = "channel2"
		is := is.New(t)

		posts, err := s.GetPosts(channelWithoutPosts, toTime(0), toTime(0))
		is.True(errors.Is(err, storage.ErrNotFound))
		is.Equal(len(posts), 0)

		_, err = s.GetLastPostID(channelWithoutPosts)
		is.True(errors.Is(err, storage.ErrNotFound))
	})
}
