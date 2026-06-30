// Package redisstate wraps Redis for the transient, real-time state the
// design calls out explicitly: live chat/vote room membership and player
// coordinates. RabbitMQ (see internal/queue) owns durable async work
// instead; Postgres (see internal/store) owns durable account/character data.
package redisstate

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"dnd5e-web/backend/internal/models"
)

type Client struct {
	rdb *redis.Client
}

func New(addr string) *Client {
	return &Client{rdb: redis.NewClient(&redis.Options{Addr: addr})}
}

func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

const coordinateTTL = 24 * time.Hour

func coordinateKey(accountID string) string {
	return fmt.Sprintf("coord:%s", accountID)
}

func (c *Client) SetCoordinate(ctx context.Context, accountID string, coord models.Coordinate) error {
	data, err := json.Marshal(coord)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, coordinateKey(accountID), data, coordinateTTL).Err()
}

func (c *Client) GetCoordinate(ctx context.Context, accountID string) (models.Coordinate, bool, error) {
	data, err := c.rdb.Get(ctx, coordinateKey(accountID)).Bytes()
	if err == redis.Nil {
		return models.Coordinate{}, false, nil
	}
	if err != nil {
		return models.Coordinate{}, false, err
	}
	var coord models.Coordinate
	if err := json.Unmarshal(data, &coord); err != nil {
		return models.Coordinate{}, false, err
	}
	return coord, true, nil
}

// Publish marshals payload to JSON and publishes it on channel — used to
// fan out chat messages, vote tallies, and dungeon-ready notifications to
// every gateway instance subscribed to that channel.
func (c *Client) Publish(ctx context.Context, channel string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.rdb.Publish(ctx, channel, data).Err()
}

func (c *Client) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return c.rdb.Subscribe(ctx, channels...)
}

// SubscribePattern subscribes to channels matching a glob pattern (e.g.
// "chat:party:*"), used for dynamically-named channels like per-party chat
// or per-job dungeon-ready notifications.
func (c *Client) SubscribePattern(ctx context.Context, pattern string) *redis.PubSub {
	return c.rdb.PSubscribe(ctx, pattern)
}

const (
	ChannelChatGlobal = "chat:global"
	ChannelChatGuild  = "chat:guild"
	ChannelChatRP     = "chat:rp"
)

func ChannelChatParty(partyID string) string {
	return fmt.Sprintf("chat:party:%s", partyID)
}

func ChannelDungeonReady(partyID string) string {
	return fmt.Sprintf("dungeon:ready:%s", partyID)
}

func ChannelVoteUpdate(promptID string) string {
	return fmt.Sprintf("vote:update:%s", promptID)
}

func ChannelVoteResolved(promptID string) string {
	return fmt.Sprintf("vote:resolved:%s", promptID)
}
