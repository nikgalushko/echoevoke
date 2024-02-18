package scrapper

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"

	"github.com/nikgalushko/echoevoke/internal/storage"
)

type ImageDownloader struct {
	db storage.ImagesStorage
}

func NewImageDownloader(db storage.ImagesStorage) *ImageDownloader {
	return &ImageDownloader{db: db}
}

func (i *ImageDownloader) DownloadImages(urls []string) []string {
	var etags []string
	for _, url := range urls {
		etag, err := i.getImageEtag(url)
		if err != nil {
			log.Println("[WARN] failed to get image etag; use uuid", err)
			etag = uuid.New().String()
		} else {
			exists, _ := i.db.IsImageExists(etag)
			if exists {
				etags = append(etags, etag)
				continue
			}
		}

		blob, err := i.getImage(url)
		if err != nil {
			log.Println("[ERROR] failed to get image", err)
			continue
		}

		if err := i.db.SaveImage(etag, blob); err != nil {
			log.Println("[ERROR] failed to save image", err)
		}

		etags = append(etags, etag)
	}

	return etags
}

func (i *ImageDownloader) getImage(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (i *ImageDownloader) getImageEtag(url string) (string, error) {
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp.Header.Get("ETag"), nil
}
