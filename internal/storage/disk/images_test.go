package disk

import (
	"context"
	"errors"
	"testing"

	"github.com/matryer/is"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nikgalushko/echoevoke/internal/storage"
)

func TestImagesStorage(t *testing.T) {
	is := is.New(t)
	ctx := context.Background()

	images := NewImagesStorage(db)

	t.Run("image does not exist", func(t *testing.T) {
		is := is.New(t)

		_, err := images.IsImageExists(ctx, "etag1")
		is.True(errors.Is(err, storage.ErrNotFound))

		data, err := images.GetImage(ctx, "etag1")
		is.True(errors.Is(err, storage.ErrNotFound))
		is.Equal(data, nil)
	})

	t.Run("save and get image", func(t *testing.T) {
		is := is.New(t)

		savedID, err := images.SaveImage(ctx, "etag1", []byte("image1"))
		is.NoErr(err)

		imageData, err := images.GetImage(ctx, "etag1")
		is.NoErr(err)
		is.Equal(string(imageData), "image1")

		existsID, err := images.IsImageExists(ctx, "etag1")
		is.NoErr(err)
		is.Equal(savedID, existsID)
	})
}
