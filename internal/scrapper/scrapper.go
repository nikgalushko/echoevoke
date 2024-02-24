package scrapper

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

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

	lastPostID, err := s.db.GetLastPostID(channelID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.Printf("[WARN] the channel %s is not registered to scape; skipped", channelID)
			return nil
		}
		return err
	}

	data, err := s.DoRequest(TMeQuery{ChannelID: channelID, LastPostID: lastPostID})
	if err != nil {
		return fmt.Errorf("failed to do the request: %w", err)
	}

	if len(data) == 0 {
		log.Println("[DEBUG] no new posts")
		return nil
	}

	log.Println("[DEBUG] parsing the page size", len(data))

	posts, err := parser.ParsePage(data)
	if err != nil {
		return fmt.Errorf("failed to parse the page: %w", err)
	}

	if len(posts) == 0 {
		log.Println("[DEBUG] no new posts")
		os.WriteFile(fmt.Sprintf("page%d.html", time.Now().Unix()), data, 0644)
		return nil
	}
	log.Println("[DEBUG] found", len(posts), "new posts")

	dbPosts := make([]storage.Post, 0, len(posts))
	for _, p := range posts {
		dbPost := storage.Post{
			Date:    p.Date,
			Message: p.Content,
			ID:      p.ID,
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

type TMeQuery struct {
	ChannelID  string
	LastPostID int64
}

func (s *Scrapper) DoRequest(query TMeQuery) ([]byte, error) {
	requestURL := "https://t.me/s/" + query.ChannelID
	log.Println("[DEBUG] request query", query)

	if query.LastPostID == 0 {
		return s.doGetRequest(requestURL)
	}

	return s.doPostRequest(requestURL, query.LastPostID)
}

func (s *Scrapper) doGetRequest(requestURL string) ([]byte, error) {
	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get the channel page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read the response body: %w", err)
	}

	return data, nil
}

func (s *Scrapper) doPostRequest(requestURL string, lastPostID int64) ([]byte, error) {
	requestURL += fmt.Sprintf("?after=%d", lastPostID)
	req, err := http.NewRequest(http.MethodPost, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create the request: %w", err)
	}

	req.Header.Set("Referer", requestURL)
	req.Header.Set("Host", "t.me")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get the channel page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var readerBody io.Reader
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create the gzip reader: %w", err)
		}
		defer gzipReader.Close()
		readerBody = gzipReader
	default:
		readerBody = resp.Body
	}

	data, err := io.ReadAll(readerBody)
	if err != nil {
		return nil, fmt.Errorf("failed to read the response body: %w", err)
	}

	data = bytes.Trim(data, `"`)
	data = bytes.ReplaceAll(data, []byte(`\"`), nil)
	data = bytes.ReplaceAll(data, []byte(`\/`), []byte("/"))

	return data, nil
}
