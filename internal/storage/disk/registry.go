package disk

import (
	"context"
	"database/sql"
	"errors"
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

func (r *ChannelRegistry) IsChannelRegistered(ctx context.Context, channelID string) (bool, error) {
	var one int
	err := r.db.QueryRowContext(ctx, "select 1 from registry where channel_id = ?", channelID).Scan(&one)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *ChannelRegistry) AllChannels(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, "select channel_id from registry")
	if err != nil {
		return nil, fmt.Errorf("failed to get all channels: %w", err)
	}
	defer rows.Close()

	var channels []string
	for rows.Next() {
		var channel string
		err = rows.Scan(&channel)
		if err != nil {
			return nil, fmt.Errorf("failed to scan channel: %w", err)
		}
		channels = append(channels, channel)
	}

	return channels, nil
}
