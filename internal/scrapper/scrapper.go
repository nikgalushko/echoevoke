package scrapper

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/nikgalushko/echoevoke/internal/parser"
	"github.com/nikgalushko/echoevoke/internal/storage"
)

type Scrapper struct {
	db   storage.PostsStorage
	imgd *ImageDownloader
}

func New(db storage.PostsStorage, imgd *ImageDownloader) *Scrapper {
	return &Scrapper{db: db, imgd: imgd}
}

func (s *Scrapper) Scrape(channelID string) error {
	latPostID, err := s.db.GetLastPostID(channelID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.Printf("[WARN] the channel %s is not registered to scape; skipped", channelID)
			return nil
		}
		return err
	}

	requestURL := "https://t.me/s/" + channelID
	if latPostID != 0 {
		requestURL += fmt.Sprintf("?before=%d", latPostID)
	}

	resp, err := http.Get(requestURL)
	if err != nil {
		return fmt.Errorf("failed to get the channel page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read the response body: %w", err)
	}

	posts, err := parser.ParsePage(data)
	if err != nil {
		return fmt.Errorf("failed to parse the page: %w", err)
	}

	if len(posts) == 0 {
		return nil
	}

	dbPosts := make([]storage.Post, 0, len(posts))
	for _, p := range posts {
		dbPost := storage.Post{
			Date:    p.Date,
			Message: p.Content,
		}
		if len(p.ImagesLink) > 0 {
			dbPost.Images = s.imgd.DownloadImages(p.ImagesLink)
		}

		dbPosts = append(dbPosts, dbPost)
	}

	return s.db.SavePosts(channelID, dbPosts)
}
