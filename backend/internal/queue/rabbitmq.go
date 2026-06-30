// Package queue wraps RabbitMQ publishing/consuming for the two
// asynchronous operations the design calls out: procedural dungeon
// generation and vote-window resolution. Both are heavy enough (grid
// generation, batched DB writes) that the gateway should hand them off
// instead of blocking a WebSocket request.
package queue

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"dnd5e-web/backend/internal/models"
)

const (
	QueueDungeonGeneration = "dungeon_generation_queue"
	QueueVoteResolution    = "vote_resolution_queue"
)

type Client struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func Connect(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}
	for _, queueName := range []string{QueueDungeonGeneration, QueueVoteResolution} {
		if _, err := ch.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("declare queue %s: %w", queueName, err)
		}
	}
	return &Client{conn: conn, ch: ch}, nil
}

func (c *Client) Close() {
	c.ch.Close()
	c.conn.Close()
}

func (c *Client) Publish(ctx context.Context, queueName string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.ch.PublishWithContext(ctx, "", queueName, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

// Consume returns a delivery channel for queueName. Acknowledge each
// delivery once processing succeeds; failed deliveries are nacked and
// requeued by the caller.
func (c *Client) Consume(queueName, consumerTag string) (<-chan amqp.Delivery, error) {
	if err := c.ch.Qos(1, 0, false); err != nil {
		return nil, fmt.Errorf("set QoS: %w", err)
	}
	return c.ch.Consume(queueName, consumerTag, false, false, false, false, nil)
}

// DungeonGenerationJob is published when an NPC quest is accepted or a
// player steps onto a [?] node.
type DungeonGenerationJob struct {
	JobID          string `json:"jobId"`
	PartyID        string `json:"partyId"`
	AccountID      string `json:"accountId"`
	CharacterLevel int    `json:"characterLevel"`
}

// VoteResolutionJob is published once a 30-second party voting window
// expires (or a solo choice resolves immediately).
type VoteResolutionJob struct {
	PromptID        string                `json:"promptId"`
	WinningOptionID string                `json:"winningOptionId"`
	Typology        models.ChoiceTypology `json:"typology"`
	AccountIDs      []string              `json:"accountIds"`
	TieBreak        bool                  `json:"tieBreak"`
}
