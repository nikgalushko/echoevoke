package scrapper

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/nikgalushko/echoevoke/internal/storage"
)

type ImageDownloader struct {
	db storage.ImagesStorage
}

func NewImageDownloader(db storage.ImagesStorage) *ImageDownloader {
	return &ImageDownloader{db: db}
}

func (i *ImageDownloader) DownloadImages(ctx context.Context, urls []string) []int64 {
	var ids []int64
	for _, url := range urls {
		etag, err := i.getImageEtag(ctx, url)
		if err != nil {
			log.Println("[WARN] failed to get image etag; use uuid", err)
			etag = uuid.New().String()
		} else {
			imgID, err := i.db.IsImageExists(ctx, etag)
			if err == nil {
				ids = append(ids, imgID)
				continue
			}
		}

		blob, err := i.getImage(ctx, url)
		if err != nil {
			log.Println("[ERROR] failed to get image", err)
			continue
		}

		imgID, err := i.db.SaveImage(ctx, etag, blob)
		if err != nil {
			log.Println("[ERROR] failed to save image", err)
		}

		ids = append(ids, imgID)
	}

	return ids
}

func (i *ImageDownloader) getImage(ctx context.Context, url string) ([]byte, error) {
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

func (i *ImageDownloader) getImageEtag(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
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

	return strings.Trim(resp.Header.Get("ETag"), `"`), nil
}
