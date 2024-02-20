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
	log.Println("[DEBUG] scraping the channel", channelID)

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
		requestURL += fmt.Sprintf("?after=%d", latPostID)
	}
	log.Println("[DEBUG] request url", requestURL)

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

	log.Println("[DEBUG] parsing the page size", len(data))

	posts, err := parser.ParsePage(data)
	if err != nil {
		return fmt.Errorf("failed to parse the page: %w", err)
	}

	if len(posts) == 0 {
		log.Println("[DEBUG] no new posts")
		return nil
	}
	log.Println("[DEBUG] found", len(posts), "new posts")

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

	err = s.db.SavePosts(channelID, dbPosts)
	if err != nil {
		return fmt.Errorf("failed to save the posts: %w", err)
	}

	return nil
}
