package sqlite

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/matryer/is"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func TestMain(m *testing.M) {
	var err error

	db, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = filepath.Walk("../../../sql", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".sql" {
			sqlBytes, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			_, err = db.ExecContext(context.Background(), string(sqlBytes))
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func TestChannelRegistry(t *testing.T) {
	const (
		knownChannel   = "channel1"
		unknownChannel = "channel2"
	)

	ctx := context.Background()
	is := is.New(t)

	registry := NewChannelRegistry(db)

	// Register a channel
	err := registry.RegisterChannel(ctx, knownChannel)
	is.NoErr(err)

	// Verify that the channel is registered
	isRegistered, err := registry.IsChannelRegistered(knownChannel)
	is.NoErr(err)
	is.True(isRegistered)

	// Unregister the channel
	err = registry.UnregisterChannel(ctx, knownChannel)
	is.NoErr(err)

	// Verify that the unregistered channel is not exist
	isRegistered, err = registry.IsChannelRegistered(knownChannel)
	is.NoErr(err)
	is.True(!isRegistered)

	// Verify that the unknown channel is not exist
	isRegistered, err = registry.IsChannelRegistered(unknownChannel)
	is.NoErr(err)
	is.True(!isRegistered)
}
