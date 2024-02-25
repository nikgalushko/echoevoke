package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ChannelRegistry struct {
	db *sql.DB
}

func NewChannelRegistry(db *sql.DB) *ChannelRegistry {
	return &ChannelRegistry{
		db: db,
	}
}

func (r *ChannelRegistry) RegisterChannel(ctx context.Context, channelID string) error {
	_, err := r.db.ExecContext(ctx,
		"insert or ignore into registry (channel_id,registered_at) values (?,?)", channelID, time.Now().UTC().Unix(),
	)
	if err != nil {
		err = fmt.Errorf("failed to register the channel: %w", err)
	}

	return err
}

func (r *ChannelRegistry) UnregisterChannel(ctx context.Context, channelID string) error {
	_, err := r.db.ExecContext(ctx, "delete from registry where channel_id = ?", channelID)
	if err != nil {
		err = fmt.Errorf("failed to unregister the channel: %w", err)
	}
	return err
}

func (r *ChannelRegistry) IsChannelRegistered(channelID string) (bool, error) {
	var one int
	err := r.db.QueryRow("select 1 from registry where channel_id = ?", channelID).Scan(&one)
	if err != nil {
		return false, err
	}
	return one == 1, nil
}
