package disk

import (
	"errors"
	"testing"

	"github.com/matryer/is"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nikgalushko/echoevoke/internal/storage"
)

func TestImagesStorage(t *testing.T) {
	is := is.New(t)

	images := NewImagesStorage(db)

	t.Run("image does not exist", func(t *testing.T) {
		is := is.New(t)

		exists, err := images.IsImageExists("etag1")
		is.NoErr(err)
		is.True(!exists)

		data, err := images.GetImage("etag1")
		is.True(errors.Is(err, storage.ErrNotFound))
		is.Equal(data, nil)
	})

	t.Run("save and get image", func(t *testing.T) {
		is := is.New(t)

		err := images.SaveImage("etag1", []byte("image1"))
		is.NoErr(err)

		imageData, err := images.GetImage("etag1")
		is.NoErr(err)
		is.Equal(string(imageData), "image1")

		exists, err := images.IsImageExists("etag1")
		is.NoErr(err)
		is.True(exists)
	})
}
