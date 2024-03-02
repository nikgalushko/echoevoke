package disk

import (
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

func (s *ImagesStorage) SaveImage(etag string, data []byte) error {
	query := "insert into images (etag, data) values (?,?)"
	_, err := s.db.Exec(query, etag, data)
	if err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	return nil
}

func (s *ImagesStorage) GetImage(etag string) ([]byte, error) {
	query := "select data from images where etag=?"
	row := s.db.QueryRow(query, etag)

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

func (s *ImagesStorage) IsImageExists(etag string) (bool, error) {
	query := "select count(*) from images where etag=?"
	row := s.db.QueryRow(query, etag)

	var count int
	err := row.Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}

		return false, fmt.Errorf("failed to check if image exists: %w", err)
	}

	return count > 0, nil
}
