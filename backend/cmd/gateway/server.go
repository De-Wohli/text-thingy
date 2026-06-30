package main

import (
	"context"
	"sync"

	"github.com/gofiber/fiber/v2"

	"dnd5e-web/backend/internal/chat"
	"dnd5e-web/backend/internal/models"
	"dnd5e-web/backend/internal/queue"
	"dnd5e-web/backend/internal/redisstate"
	"dnd5e-web/backend/internal/store"
	"dnd5e-web/backend/internal/voting"
	"dnd5e-web/backend/internal/wsproto"
)

type server struct {
	store *store.Store
	redis *redisstate.Client
	queue *queue.Client
	hub   *chat.Hub

	voteRoomsMu sync.Mutex
	voteRooms   map[string]*voting.VoteRoom // promptID -> room

	dungeonsMu sync.Mutex
	dungeons   map[string]*models.Dungeon // partyKey -> in-flight dungeon
}

func newServer(st *store.Store, rs *redisstate.Client, q *queue.Client) *server {
	return &server{
		store:     st,
		redis:     rs,
		queue:     q,
		hub:       chat.NewHub(),
		voteRooms: make(map[string]*voting.VoteRoom),
		dungeons:  make(map[string]*models.Dungeon),
	}
}

// partyKey returns the account's party ID, or the account ID itself when
// solo — this lets dungeon/voting code treat "solo" as a party of one
// instead of branching everywhere.
func partyKey(account models.Account) string {
	if account.PartyID != nil && *account.PartyID != "" {
		return *account.PartyID
	}
	return account.ID
}

func (s *server) stateSync(ctx context.Context, accountID string) (wsproto.StateSync, error) {
	account, err := s.store.GetAccount(ctx, accountID)
	if err != nil {
		return wsproto.StateSync{}, err
	}
	characters, err := s.store.ListCharacters(ctx, accountID)
	if err != nil {
		return wsproto.StateSync{}, err
	}
	return wsproto.NewStateSync(account, characters), nil
}

// --- REST handlers ---

type createAccountRequest struct {
	DisplayName string `json:"displayName"`
}

func (s *server) handleCreateAccount(c *fiber.Ctx) error {
	var req createAccountRequest
	if err := c.BodyParser(&req); err != nil || req.DisplayName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "displayName is required"})
	}
	account, err := s.store.CreateAccount(c.Context(), req.DisplayName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(account)
}

func (s *server) handleGetAccount(c *fiber.Ctx) error {
	sync, err := s.stateSync(c.Context(), c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "account not found"})
	}
	return c.JSON(sync)
}
