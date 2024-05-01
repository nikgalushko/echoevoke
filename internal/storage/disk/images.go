package disk

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/nikgalushko/echoevoke/internal/storage"
)

type ImagesStorage struct {
	db *sql.DB
}

func NewImagesStorage(db *sql.DB) *ImagesStorage {
	return &ImagesStorage{
		db: db,
	}
}

func (s *ImagesStorage) SaveImage(ctx context.Context, etag string, data []byte) (int64, error) {
	var imgID int64
	query := "insert into images (etag, data) values (?,?) returning id"
	err := s.db.QueryRowContext(ctx, query, etag, data).Scan(&imgID)
	if err != nil {
		return 0, fmt.Errorf("failed to save image: %w", err)
	}

	return imgID, nil
}

func (s *ImagesStorage) GetImage(ctx context.Context, etag string) ([]byte, error) {
	query := "select data from images where etag=?"
	row := s.db.QueryRowContext(ctx, query, etag)

	var data []byte
	err := row.Scan(&data)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	return data, nil
}

// TODO: либо etag либо id сейчас какая-то херня; storage.Post хранит id, а получаем по etag
func (s *ImagesStorage) GetImageByID(ctx context.Context, id int64) ([]byte, error) {
	query := "select data from images where id=?"
	row := s.db.QueryRowContext(ctx, query, id)

	var data []byte
	err := row.Scan(&data)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	return data, nil
}

func (s *ImagesStorage) IsImageExists(ctx context.Context, etag string) (int64, error) {
	query := "select id from images where etag=?"
	row := s.db.QueryRowContext(ctx, query, etag)

	var imgID int64
	err := row.Scan(&imgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, storage.ErrNotFound
		}

		return 0, fmt.Errorf("failed to check if image exists: %w", err)
	}

	return imgID, nil
}
