package disk

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nikgalushko/echoevoke/internal/storage"
)

type PostsStorage struct {
	db *sql.DB
}

func NewPostsStorage(db *sql.DB) *PostsStorage {
	return &PostsStorage{
		db: db,
	}
}

func (s *PostsStorage) SavePosts(channelID string, posts []storage.Post) (err error) {
	var tx *sql.Tx
	tx, err = s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	var postStmt, imageStmt *sql.Stmt

	postStmt, err = tx.Prepare("insert into posts (id, channel_id, date, message) values (?,?,?,?)")
	if err != nil {
		return fmt.Errorf("failed to prepare post statement: %w", err)
	}
	defer postStmt.Close()

	imageStmt, err = tx.Prepare("insert into post_images (post_id, image_id) values (?,?)")
	if err != nil {
		return fmt.Errorf("failed to prepare image statement: %w", err)
	}
	defer imageStmt.Close()

	for _, post := range posts {
		_, err = postStmt.Exec(post.ID, channelID, post.Date.UTC().Unix(), post.Message)
		if err != nil {
			return fmt.Errorf("failed to save post: %w", err)
		}

		for _, imageID := range post.Images {
			_, err = imageStmt.Exec(post.ID, imageID)
			if err != nil {
				return fmt.Errorf("failed to save image: %w", err)
			}
		}
	}

	return nil
}

func (s *PostsStorage) GetPosts(channelID string, from, to time.Time) ([]storage.Post, error) {
	rows, err := s.db.Query("select id, date, message from posts where channel_id=? and date >= ? and date < ? order by id asc",
		channelID, from.UTC().Unix(), to.UTC().Unix(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}

	var (
		posts []storage.Post
		ids   []any
	)
	for rows.Next() {
		var (
			post          storage.Post
			unixTimestamp int64
		)
		err := rows.Scan(&post.ID, &unixTimestamp, &post.Message)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		post.Date = time.Unix(unixTimestamp, 0).UTC()

		posts = append(posts, post)
		ids = append(ids, post.ID)
	}

	if len(posts) == 0 {
		return nil, storage.ErrNotFound
	}

	if len(ids) == 0 {
		return posts, nil
	}

	rows, err = s.db.Query(`select posts.id as post_id, post_images.image_id from posts
		join post_images on post_images.post_id = posts.id
		where posts.id in (?`+strings.Repeat(",?", len(ids)-1)+`)
		order by post_id asc`,
		ids...,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return posts, nil
		}

		return nil, fmt.Errorf("failed to get images: %w", err)
	}

	prevPostID := int64(-1)
	i := 0
	for rows.Next() {
		var imageID, postID int64

		err := rows.Scan(&postID, &imageID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan image: %w", err)
		}

		if prevPostID != postID {
			prevPostID = postID
			for posts[i].ID != postID {
				i++
			}
		}

		posts[i].Images = append(posts[i].Images, imageID)
	}

	return posts, nil
}

func (s *PostsStorage) GetLastPost(channelID string) (storage.Post, error) {
	return storage.Post{}, errors.New("not implemented")
}

func (s *PostsStorage) GetLastPostID(channelID string) (int64, error) {
	var lastPostID int64
	err := s.db.QueryRow("select id from posts where channel_id=? order by id desc limit 1", channelID).Scan(&lastPostID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, storage.ErrNotFound
		}

		return 0, fmt.Errorf("failed to get last post id: %w", err)
	}

	return lastPostID, nil
}
